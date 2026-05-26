package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"rtm-server/internal/constant"
	"rtm-server/internal/model"
	"rtm-server/internal/utils"
)

// ChatRepository Agent聊天会话/消息数据访问
type ChatRepository struct {
	db *gorm.DB
}

// NewChatRepository 创建聊天仓库
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CountSessions 统计指定用户 + 会话类型 的会话总数
func (r *ChatRepository) CountSessions(ctx context.Context, userID int64, sessionType string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.ChatSession{}).
		Where("user_id = ? AND session_type = ? AND del_flag = ?", userID, sessionType, constant.DelFlagExist).
		Count(&total).Error
	return total, err
}

// FindSessionPage 分页查询指定用户 + 会话类型 的会话列表（按 last_message_at DESC, id DESC）
func (r *ChatRepository) FindSessionPage(ctx context.Context, userID int64, sessionType string, offset, limit int) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND session_type = ? AND del_flag = ?", userID, sessionType, constant.DelFlagExist).
		Order("last_message_at DESC NULLS LAST, id DESC").
		Offset(offset).Limit(limit).
		Find(&sessions).Error
	return sessions, err
}

// FindSessionByID 按 ID 查询会话（带 user_id 隔离）
func (r *ChatRepository) FindSessionByID(ctx context.Context, userID, id int64) (*model.ChatSession, error) {
	var s model.ChatSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND del_flag = ?", id, userID, constant.DelFlagExist).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindMessagesBySession 查询会话内全部消息（按 seq 升序）
func (r *ChatRepository) FindMessagesBySession(ctx context.Context, sessionID int64) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("seq ASC, id ASC").
		Find(&messages).Error
	return messages, err
}

// AppendTurnInput AppendTurn 入参
type AppendTurnInput struct {
	UserID           int64
	Username         string
	SessionID        int64  // 0 表示新建会话
	SessionType      string // 仅在 SessionID == 0 时使用：control / perceive
	InitialTitle     string // 仅在 SessionID == 0 时使用
	AgentSessionID   string
	UserMessage      string
	AssistantMessage string
}

// AppendTurn 在单事务中创建/获取会话，并追加 user/assistant 两条消息
// 返回最新的 session 记录
func (r *ChatRepository) AppendTurn(ctx context.Context, in *AppendTurnInput) (*model.ChatSession, error) {
	now := utils.NowStr()
	var session model.ChatSession

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if in.SessionID == 0 {
			session = model.ChatSession{
				UserID:         in.UserID,
				SessionType:    in.SessionType,
				Title:          in.InitialTitle,
				AgentSessionID: in.AgentSessionID,
				LastMessageAt:  now,
				MessageCount:   0,
				BaseEntity: model.BaseEntity{
					CreateBy: in.Username,
				},
			}
			if err := tx.Create(&session).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("id = ? AND user_id = ? AND del_flag = ?", in.SessionID, in.UserID, constant.DelFlagExist).
				First(&session).Error; err != nil {
				return err
			}
		}

		// 取当前 max(seq)
		var maxSeq int
		if err := tx.Model(&model.ChatMessage{}).
			Where("session_id = ?", session.ID).
			Select("COALESCE(MAX(seq), 0)").
			Scan(&maxSeq).Error; err != nil {
			return err
		}

		// 批量写入 user + assistant 两条消息
		msgs := []model.ChatMessage{
			{
				SessionID:  session.ID,
				Role:       constant.ChatRoleUser,
				Content:    in.UserMessage,
				Seq:        maxSeq + 1,
				CreateTime: now,
			},
			{
				SessionID:  session.ID,
				Role:       constant.ChatRoleAssistant,
				Content:    in.AssistantMessage,
				Seq:        maxSeq + 2,
				CreateTime: now,
			},
		}
		if err := tx.Create(&msgs).Error; err != nil {
			return err
		}

		newCount := maxSeq + 2
		updates := map[string]interface{}{
			"last_message_at": now,
			"message_count":   newCount,
			"update_by":       in.Username,
		}
		// 仅在 agentSessionId 非空且发生变化时更新，避免覆盖已有值
		if in.AgentSessionID != "" && in.AgentSessionID != session.AgentSessionID {
			updates["agent_session_id"] = in.AgentSessionID
			session.AgentSessionID = in.AgentSessionID
		}
		if err := tx.Model(&model.ChatSession{}).
			Where("id = ?", session.ID).
			Updates(updates).Error; err != nil {
			return err
		}

		// 同步内存中的 session 字段，便于返回
		session.LastMessageAt = now
		session.MessageCount = newCount
		session.UpdateBy = in.Username
		session.UpdateTime = now
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Rename 修改会话标题（带 user_id 隔离）
func (r *ChatRepository) Rename(ctx context.Context, userID, id int64, title, updateBy string) error {
	res := r.db.WithContext(ctx).Model(&model.ChatSession{}).
		Where("id = ? AND user_id = ? AND del_flag = ?", id, userID, constant.DelFlagExist).
		Updates(map[string]interface{}{
			"title":     title,
			"update_by": updateBy,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("会话不存在或无权限")
	}
	return nil
}

// DeleteByIDs 批量删除会话：会话软删（del_flag='2'），消息硬删（同事务）
// 仅删除属于 userID 的会话
func (r *ChatRepository) DeleteByIDs(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先筛选出实际属于该用户的会话 ID（防止跨用户删除）
		var ownedIDs []int64
		if err := tx.Model(&model.ChatSession{}).
			Where("id IN ? AND user_id = ? AND del_flag = ?", ids, userID, constant.DelFlagExist).
			Pluck("id", &ownedIDs).Error; err != nil {
			return err
		}
		if len(ownedIDs) == 0 {
			return nil
		}

		// 硬删消息
		if err := tx.Where("session_id IN ?", ownedIDs).
			Delete(&model.ChatMessage{}).Error; err != nil {
			return err
		}

		// 软删会话
		if err := tx.Model(&model.ChatSession{}).
			Where("id IN ?", ownedIDs).
			Update("del_flag", constant.DelFlagDelete).Error; err != nil {
			return err
		}
		return nil
	})
}
