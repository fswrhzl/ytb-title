// 频道相关操作
package gorm

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ChannelRepository interface {
	// 获取所有频道
	GetAllChannels() ([]*ChannelResponse, error)
	// 创建频道
	CreateChannel(ccr *ChannelCreateRequest) error
	// 更新频道
	UpdateChannel(cur *ChannelUpdateRequest) error
	// 删除频道
	DeleteChannel(id int) error
}

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository() ChannelRepository {
	return &channelRepository{db: DB}
}

func (r *channelRepository) GetAllChannels() ([]*ChannelResponse, error) {
	var channels []*ChannelResponse
	rows, err := r.db.Table("channels AS c").
		Select("c.id, c.name, GROUP_CONCAT(ct.tag_id, ',') AS tagListStr").
		Joins(" left join channel_tag AS ct on c.id = ct.channel_id").
		Group("c.id").
		Rows()
	if err != nil {
		log.Printf("查询所有频道失败：%v", err)
		return nil, errors.New("查询所有频道失败")
	}
	defer rows.Close()
	for rows.Next() {
		var channel ChannelResponse
		// 数据库中的null不对应任何go中的数据类型，需要特殊处理，使用sql.NullString类型接收
		var tagListStrTmp sql.NullString
		if err := rows.Scan(&channel.Id, &channel.Name, &tagListStrTmp); err != nil {
			return nil, errors.New("数据解析失败")
		}
		var tagListStr string
		if tagListStrTmp.Valid { // 如果不为null
			tagListStr = tagListStrTmp.String
		} else { // 如果为null，转换为空字符串
			tagListStr = ""
		}
		for len(tagListStr) > 0 {
			tagId, rest, found := strings.Cut(tagListStr, ",")
			tagIdInt, err := strconv.Atoi(tagId)
			if err != nil {
				return nil, errors.New("标签ID转换失败")
			}
			channel.Tags = append(channel.Tags, int64(tagIdInt))
			if !found {
				break
			}
			tagListStr = rest
		}
		channels = append(channels, &channel)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("迭代频道行失败")
	}
	return channels, nil
}

func (r *channelRepository) CreateChannel(ccr *ChannelCreateRequest) error {
	var channel Channel = Channel{Name: ccr.Name}
	result := r.db.Create(&channel)
	if result.Error != nil {
		log.Printf("插入频道失败：%v", result.Error)
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return errors.New("频道名称已存在")
		}
		return errors.New("插入频道失败")
	}
	for _, tagId := range ccr.Tags {
		result := r.db.Create(&ChannelTag{ChannelId: channel.Id, TagId: tagId})
		if result.Error != nil {
			log.Printf("插入频道标签失败：%v", result.Error)
			return errors.New("插入频道标签失败")
		}
	}
	return nil
}

func (r *channelRepository) UpdateChannel(cur *ChannelUpdateRequest) error {
	var channel Channel = Channel{Id: cur.Id, Name: cur.Name}
	// Save方法默认使用id作为条件，更新其他字段
	result := r.db.Save(&channel)
	if result.Error != nil {
		log.Printf("更新频道失败：%v", result.Error)
		return errors.New("更新频道失败")
	}
	result = r.db.Delete(&ChannelTag{}, "channel_id = ?", cur.Id)
	if result.Error != nil {
		log.Printf("删除频道标签失败：%v", result.Error)
		return errors.New("删除频道标签失败")
	}
	for _, tagId := range cur.Tags {
		result := r.db.Create(&ChannelTag{ChannelId: cur.Id, TagId: tagId})
		if result.Error != nil {
			log.Printf("插入频道标签失败：%v", result.Error)
			return errors.New("插入频道标签失败")
		}
	}
	return nil
}

func (r *channelRepository) DeleteChannel(id int) error {
	result := r.db.Delete(&Channel{}, id)
	if result.Error != nil {
		log.Printf("删除频道失败：%v", result.Error)
		return errors.New("删除频道失败")
	}
	result = r.db.Delete(&ChannelTag{}, "channel_id = ?", id)
	if result.Error != nil {
		log.Printf("删除频道标签失败：%v", result.Error)
		return errors.New("删除频道标签失败")
	}
	return nil
}
