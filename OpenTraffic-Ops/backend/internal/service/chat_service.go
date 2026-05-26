package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// titleMaxRunes 标题最大字符数（rune）
const titleMaxRunes = 24

// renameMaxRunes 重命名标题最大字符数（rune）
const renameMaxRunes = 128

// ChatService Agent聊天会话服务
type ChatService struct {
	repo *repository.ChatRepository
}

// NewChatService 创建聊天服务
func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{repo: repository.NewChatRepository(db)}
}

// DeriveTitle 由首条用户消息派生标题：去首尾空白、合并空白、最多 24 个 rune，超出加省略号
func DeriveTitle(msg string) string {
	s := strings.TrimSpace(msg)
	s = strings.NewReplacer("\r\n", " ", "\n", " ", "\t", " ").Replace(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	if s == "" {
		return "新对话"
	}
	if len([]rune(s)) > titleMaxRunes {
		return utils.SubString(s, 0, titleMaxRunes) + "…"
	}
	return s
}

// ListSessions 分页查询当前用户在指定类型下的会话列表
func (s *ChatService) ListSessions(ctx context.Context, userID int64, sessionType string, query *dto.ChatSessionQuery) ([]dto.ChatSessionDto, int64, error) {
	total, err := s.repo.CountSessions(ctx, userID, sessionType)
	if err != nil {
		return nil, 0, err
	}

	sessions, err := s.repo.FindSessionPage(ctx, userID, sessionType, query.GetOffset(), query.GetLimit())
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.ChatSessionDto, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, dto.ChatSessionDto{
			ID:             sess.ID,
			SessionType:    sess.SessionType,
			Title:          sess.Title,
			AgentSessionID: sess.AgentSessionID,
			LastMessageAt:  sess.LastMessageAt,
			MessageCount:   sess.MessageCount,
			CreateTime:     sess.CreateTime,
			UpdateTime:     sess.UpdateTime,
		})
	}
	return result, total, nil
}

// GetSessionDetail 获取指定会话的详情（含消息）
func (s *ChatService) GetSessionDetail(ctx context.Context, userID, id int64) (*dto.ChatSessionDetailDto, error) {
	sess, err := s.repo.FindSessionByID(ctx, userID, id)
	if err != nil {
		return nil, errors.New("会话不存在")
	}

	messages, err := s.repo.FindMessagesBySession(ctx, sess.ID)
	if err != nil {
		return nil, err
	}

	msgs := make([]dto.ChatMessageDto, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, dto.ChatMessageDto{
			ID:         m.ID,
			Role:       m.Role,
			Content:    m.Content,
			Seq:        m.Seq,
			CreateTime: m.CreateTime,
		})
	}

	return &dto.ChatSessionDetailDto{
		ChatSessionDto: dto.ChatSessionDto{
			ID:             sess.ID,
			SessionType:    sess.SessionType,
			Title:          sess.Title,
			AgentSessionID: sess.AgentSessionID,
			LastMessageAt:  sess.LastMessageAt,
			MessageCount:   sess.MessageCount,
			CreateTime:     sess.CreateTime,
			UpdateTime:     sess.UpdateTime,
		},
		Messages: msgs,
	}, nil
}

// SaveTurn 保存一轮 user/assistant 对话；sessionId 为 0 时新建会话并使用派生标题
// sessionType 仅在新建会话（sessionId==0）时使用
func (s *ChatService) SaveTurn(ctx context.Context, userID int64, username, sessionType string, req *dto.ChatTurnRequest) (*dto.ChatTurnResponse, error) {
	in := &repository.AppendTurnInput{
		UserID:           userID,
		Username:         username,
		SessionID:        req.SessionID,
		SessionType:      sessionType,
		AgentSessionID:   strings.TrimSpace(req.AgentSessionID),
		UserMessage:      req.UserMessage,
		AssistantMessage: req.AssistantMessage,
	}
	if req.SessionID == 0 {
		in.InitialTitle = DeriveTitle(req.UserMessage)
	}

	sess, err := s.repo.AppendTurn(ctx, in)
	if err != nil {
		return nil, err
	}

	return &dto.ChatTurnResponse{
		SessionID:     sess.ID,
		Title:         sess.Title,
		LastMessageAt: sess.LastMessageAt,
		MessageCount:  sess.MessageCount,
	}, nil
}

// Rename 重命名会话
func (s *ChatService) Rename(ctx context.Context, userID int64, req *dto.ChatSessionRenameRequest, updateBy string) error {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errors.New("标题不能为空")
	}
	if len([]rune(title)) > renameMaxRunes {
		title = utils.SubString(title, 0, renameMaxRunes)
	}
	return s.repo.Rename(ctx, userID, req.ID, title, updateBy)
}

// DeleteByIDs 批量删除会话（仅删当前用户的）
func (s *ChatService) DeleteByIDs(ctx context.Context, userID int64, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, userID, ids)
}
