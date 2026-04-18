package grpcserver

import (
	"time"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// timeProto преобразует time.Time в timestamppb.Timestamp.
func timeProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// protoSecretVersion преобразует model.SecretVersion в gophkeeperpb.SecretVersion.
func protoSecretVersion(v *model.SecretVersion) *gophkeeperpb.SecretVersion {
	if v == nil {
		return nil
	}
	return &gophkeeperpb.SecretVersion{
		Id:                v.ID,
		SecretId:          v.SecretID,
		Version:           int32(v.Version),
		DataFormatVersion: int32(v.DataFormatVersion),
		DataEncrypted:     v.DataEncrypted,
		DataSize:          int32(v.DataSize),
		CreatedAt:         timeProto(*v.CreatedAt),
		UpdatedAt:         timeProto(*v.UpdatedAt),
		DeletedAt:         timeProto(*v.DeletedAt),
	}
}

// protoSecret преобразует model.Secret в gophkeeperpb.Secret.
func protoSecret(s *model.Secret, current *model.SecretVersion, history []*model.SecretVersion, includeHistory bool) *gophkeeperpb.Secret {
	if s == nil {
		return nil
	}
	out := &gophkeeperpb.Secret{
		Id:                     s.ID,
		UserId:                 s.UserID,
		CurrentSecretVersionId: s.CurrentSecretVersionID,
		CreatedAt:              timeProto(*s.CreatedAt),
		UpdatedAt:              timeProto(*s.UpdatedAt),
		DeletedAt:              timeProto(*s.DeletedAt),
		CurrentVersion:         protoSecretVersion(current),
	}
	if includeHistory && len(history) > 0 {
		vers := make([]*gophkeeperpb.SecretVersion, 0, len(history))
		for _, h := range history {
			vers = append(vers, protoSecretVersion(h))
		}
		out.SecretVersions = vers
	}
	return out
}

// protoAttachmentSummary преобразует model.AttachmentSummary в gophkeeperpb.AttachmentSummary.
func protoAttachmentSummary(a *model.AttachmentSummary) *gophkeeperpb.AttachmentSummary {
	if a == nil {
		return nil
	}
	return &gophkeeperpb.AttachmentSummary{
		Id:                a.ID,
		SecretVersionId:   a.SecretVersionID,
		DataFormatVersion: int32(a.DataFormatVersion),
		InfoFormatVersion: int32(a.InfoFormatVersion),
		InfoEncrypted:     a.InfoEncrypted,
		InfoSize:          int32(a.InfoSize),
		DataSize:          int32(a.DataSize),
		CreatedAt:         timeProto(*a.CreatedAt),
	}
}

// protoAttachment преобразует model.Attachment в gophkeeperpb.Attachment.
func protoAttachment(a *model.Attachment) *gophkeeperpb.Attachment {
	if a == nil {
		return nil
	}
	return &gophkeeperpb.Attachment{
		Id:                a.ID,
		SecretVersionId:   a.SecretVersionID,
		DataFormatVersion: int32(a.DataFormatVersion),
		InfoFormatVersion: int32(a.InfoFormatVersion),
		InfoEncrypted:     a.InfoEncrypted,
		InfoSize:          int32(a.InfoSize),
		DataEncrypted:     a.DataEncrypted,
		DataSize:          int32(a.DataSize),
		CreatedAt:         timeProto(*a.CreatedAt),
	}
}
