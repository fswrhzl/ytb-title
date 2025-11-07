// 频道相关操作
package db

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
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
	db *sql.DB
}

func NewChannelRepository() ChannelRepository {
	return &channelRepository{db: DB}
}

func (r *channelRepository) GetAllChannels() ([]*ChannelResponse, error) {
	var channels []*ChannelResponse
	rows, err := r.db.Query(
		`SELECT
			c.id   AS cid,
			c.name AS cname,
			GROUP_CONCAT(ct.tag_id, ',') AS tagListStr
		FROM channels AS c
		LEFT JOIN channel_tag AS ct ON c.id = ct.channel_id
		GROUP BY c.id`)
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
	query := "insert into channels (name) values (?)"
	result, err := r.db.Exec(query, ccr.Name)
	if err != nil {
		log.Printf("插入频道失败：%v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("频道名称已存在")
		}
		return errors.New("插入频道失败")
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("获取插入的频道ID失败：%v", err)
		return errors.New("获取插入的频道ID失败")
	}
	query = "insert into channel_tag (channel_id, tag_id) values (?, ?)"
	for _, tagId := range ccr.Tags {
		if _, err := r.db.Exec(query, id, tagId); err != nil {
			log.Printf("插入频道标签失败：%v", err)
			return errors.New("插入频道标签失败")
		}
	}
	return nil
}

func (r *channelRepository) UpdateChannel(cur *ChannelUpdateRequest) error {
	query := "update channels set name = ? where id = ?"
	if _, err := r.db.Exec(query, cur.Name, cur.Id); err != nil {
		log.Printf("更新频道失败：%v", err)
		return errors.New("更新频道失败")
	}
	query = "delete from channel_tag where channel_id = ?"
	if _, err := r.db.Exec(query, cur.Id); err != nil {
		log.Printf("删除频道标签失败：%v", err)
		return errors.New("删除频道标签失败")
	}
	query = "insert into channel_tag (channel_id, tag_id) values (?, ?)"
	for _, tagId := range cur.Tags {
		if _, err := r.db.Exec(query, cur.Id, tagId); err != nil {
			log.Printf("插入频道标签失败：%v", err)
			return errors.New("插入频道标签失败")
		}
	}
	return nil
}

func (r *channelRepository) DeleteChannel(id int) error {
	query := "delete from channels where id = ?"
	if _, err := r.db.Exec(query, id); err != nil {
		log.Printf("删除频道失败：%v", err)
		return errors.New("删除频道失败")
	}
	query = "delete from channel_tag where channel_id = ?"
	if _, err := r.db.Exec(query, id); err != nil {
		log.Printf("删除频道标签失败：%v", err)
		return errors.New("删除频道标签失败")
	}
	return nil
}
