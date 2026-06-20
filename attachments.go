package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// AttachmentClient uploads, inspects, and downloads attachments. Obtain it from
// [Client.Attachments].
type AttachmentClient struct {
	svc imessagev1connect.AttachmentServiceClient
}

// Get returns metadata for an attachment without transferring its bytes.
func (a *AttachmentClient) Get(ctx context.Context, attachmentGUID string) (AttachmentInfo, error) {
	req := &imessagev1.GetAttachmentInfoRequest{}
	req.SetAttachmentGuid(attachmentGUID)
	resp, err := a.svc.GetAttachmentInfo(ctx, connect.NewRequest(req))
	if err != nil {
		return AttachmentInfo{}, asError(err)
	}
	return attachmentInfoFromProto(resp.Msg.GetAttachment()), nil
}

// Upload stores a file (and optional Live Photo companion) and returns its
// metadata. The primary and companion are persisted atomically.
func (a *AttachmentClient) Upload(ctx context.Context, input AttachmentInput) (UploadAttachmentResult, error) {
	req := &imessagev1.UploadAttachmentRequest{}
	req.SetFileName(input.FileName)
	req.SetData(input.Data)
	if input.Companion != nil {
		comp := &imessagev1.Companion{}
		comp.SetData(input.Companion.Data)
		comp.SetKind(companionKindToProto(input.Companion.Kind))
		req.SetCompanion(comp)
	}
	resp, err := a.svc.UploadAttachment(ctx, connect.NewRequest(req))
	if err != nil {
		return UploadAttachmentResult{}, asError(err)
	}
	res := UploadAttachmentResult{Attachment: attachmentInfoFromProto(resp.Msg.GetAttachment())}
	if resp.Msg.HasCompanion() {
		res.Companion = companionInfoFromProto(resp.Msg.GetCompanion())
	}
	return res, nil
}

// DownloadStream streams an attachment's bytes. The first frame is a
// [DownloadHeader]; subsequent frames are [DownloadPrimaryChunk] and, for a
// companion, [DownloadCompanionChunk]. Always Close the returned stream.
func (a *AttachmentClient) DownloadStream(ctx context.Context, attachmentGUID string) *DownloadStream {
	req := &imessagev1.DownloadAttachmentRequest{}
	req.SetAttachmentGuid(attachmentGUID)
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.DownloadAttachmentResponse], error) {
			return a.svc.DownloadAttachment(ctx, connect.NewRequest(req))
		},
		downloadChunkFromProto,
	)
	return &DownloadStream{stream[DownloadChunk]{src: sub}}
}
