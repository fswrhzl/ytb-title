// 标签数据操作
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type TagRepository interface {
	CreateTag(tcr *TagCreateRequest) error
	DeleteTag(id int) error
	ListTags() ([]*TagResponse, error)
}

type tagRepository struct {
	db *sql.DB
}

func NewTagRepository() TagRepository {
	return &tagRepository{db: DB}
}

func (tr *tagRepository) CreateTag(tcr *TagCreateRequest) error {
	// 新增标签
	query := "INSERT INTO tags (name) VALUES (?)"
	result, err := tr.db.Exec(query, tcr.Name)
	if err != nil {
		log.Printf("创建标签失败: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("标签名已存在")
		}
		return errors.New("创建标签失败")
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("获取标签ID失败: %v", err)
		return errors.New("获取标签ID失败")
	}
	// 新增标签与频道的关联关系
	query = "INSERT INTO channel_tag (channel_id, tag_id) VALUES (?, ?)"
	for _, channelId := range tcr.Channels {
		_, err = tr.db.Exec(query, channelId, id)
		if err != nil {
			log.Printf("创建标签与频道关联关系失败: %v", err)
			return errors.New("创建标签与频道关联关系失败")
		}
	}

	return nil
}

func (tr *tagRepository) DeleteTag(id int) error {
	query := "DELETE FROM tags WHERE id = ?"
	result, err := tr.db.Exec(query, id)
	if err != nil {
		log.Printf("删除标签失败: %v", err)
		return errors.New("删除标签失败")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除标签影响行数失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("未发现该标签: %d", id)
	}
	query = "DELETE FROM channel_tag WHERE tag_id = ?"
	_, err = tr.db.Exec(query, id)
	if err != nil {
		log.Printf("删除标签与频道关联关系失败: %v", err)
		return errors.New("删除标签与频道关联关系失败")
	}
	return nil
}

func (tr *tagRepository) ListTags() ([]*TagResponse, error) {
	query := `
		SELECT
			t.id   AS tid,
			t.name AS tname,
			GROUP_CONCAT(c.channel_id, ',') AS tlink
		FROM tags AS t
		LEFT JOIN channel_tag AS c
			ON t.id = c.tag_id
		GROUP BY tid;
	`
	rows, err := tr.db.Query(query)
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
