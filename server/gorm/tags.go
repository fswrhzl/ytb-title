// 标签数据操作
package gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type TagRepository interface {
	CreateTag(tcr *TagCreateRequest) error
	DeleteTag(id int) error
	ListTags() ([]*TagResponse, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository() TagRepository {
	return &tagRepository{db: DB}
}

func (tr *tagRepository) CreateTag(tcr *TagCreateRequest) error {
	// 新增标签
	var tag Tag = Tag{
		Name: tcr.Name,
	}
	err := tr.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tag).Error; err != nil {
			log.Printf("创建标签失败: %v", err)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return errors.New("标签名已存在")
			}
			return errors.New("创建标签失败")
		}
		fmt.Printf("创建标签成功: %v", tag)
		// 新增标签与频道的关联关系
		for _, channelId := range tcr.Channels {
			ctLink := ChannelTag{
				ChannelId: channelId,
				TagId:     tag.Id,
			}
			if err := tx.Create(&ctLink).Error; err != nil {
				log.Printf("为标签设置关联频道失败: %v", err)
				return errors.New("为标签设置关联频道失败")
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("创建标签时，开启事务失败：%v", err.Error())
		return err
	}
	return nil
}

func (tr *tagRepository) DeleteTag(id int) error {
	err := tr.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&Tag{}, id).Error
		if err != nil {
			log.Printf("删除标签失败: %v", err)
			return errors.New("删除标签失败")
		}

		err = tx.Delete(&ChannelTag{}, "tag_id = ?", id).Error
		if err != nil {
			log.Printf("删除标签与频道关联关系失败: %v", err)
			return errors.New("删除标签与频道关联关系失败")
		}
		return nil
	})
	if err != nil {
		log.Printf("删除标签时，开启事务失败：%v", err.Error())
		return errors.New("删除标签失败")
	}
	return nil
}

func (tr *tagRepository) ListTags() ([]*TagResponse, error) {
	rows, err := tr.db.Table("tags as t").
		Select("t.id, t.name, GROUP_CONCAT(c.channel_id, ',') AS tlink").
		Joins("LEFT JOIN channel_tag AS c ON t.id = c.tag_id").
		Group("t.id").
		Rows()
	if err != nil {
		log.Printf("查询标签失败: %v", err)
		return nil, errors.New("查询标签失败")
	}
	defer rows.Close()

	var tagListResponse []*TagResponse
	for rows.Next() {
		// 理论上循环体内使用var声明的变量，在每次迭代时都是一个新的变量，不会互相影响。绝大多数循环的临时变量都保存在栈上。在不逃逸的情况下，栈空间可能会被复用。
		// 循环中声明的变量在本次循环开始时初始化，在本次循环结束时销毁，避免内存泄漏。
		// 但是如果变量的地址被保存到循环体外，变量会逃逸到堆上，下次循环开始时声明的变量会是新的实例，重新分配内存空间
		var tag TagResponse
		var channelStrTmp sql.NullString
		var channelStr string
		if err := rows.Scan(&tag.Id, &tag.Name, &channelStrTmp); err != nil {
			return nil, err
		}
		if channelStrTmp.Valid {
			channelStr = channelStrTmp.String
		} else {
			channelStr = ""
		}
		for len(channelStr) > 0 {
			channelId, rest, found := strings.Cut(channelStr, ",")
			channelIdInt, err := strconv.Atoi(channelId)
			if err != nil {
				log.Printf("转换频道ID失败: %v", err)
				return nil, errors.New("转换频道ID失败")
			}
			tag.Channels = append(tag.Channels, int64(channelIdInt))
			if !found {
				break
			}
			channelStr = rest
		}
		tagListResponse = append(tagListResponse, &tag) // 使用指针避免值拷贝，确保切片中存储的是同一份 tag 实例，节省内存并保证后续若修改 tag 会反映到切片中
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("查询标签失败")
	}
	return tagListResponse, nil
}
