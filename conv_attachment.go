package imessage

import (
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

func attachmentInfoFromProto(pb *imessagev1.AttachmentInfo) AttachmentInfo {
	if pb == nil {
		return AttachmentInfo{}
	}
	ai := AttachmentInfo{
		GUID:          pb.GetGuid(),
		OriginalGUID:  pb.GetOriginalGuid(),
		FileName:      pb.GetFileName(),
		MIMEType:      pb.GetMimeType(),
		UTI:           pb.GetUti(),
		TotalBytes:    pb.GetTotalBytes(),
		IsOutgoing:    pb.GetIsOutgoing(),
		TransferState: transferStateFromProto(pb.GetTransferState()),
		IsHidden:      pb.GetIsHidden(),
		IsSticker:     pb.GetIsSticker(),
	}
	if pb.HasCompanionKind() {
		ai.CompanionKind = companionKindFromProto(pb.GetCompanionKind())
	}
	return ai
}

func attachmentInfoPtrFromProto(pb *imessagev1.AttachmentInfo) *AttachmentInfo {
	if pb == nil {
		return nil
	}
	a := attachmentInfoFromProto(pb)
	return &a
}

func companionInfoFromProto(pb *imessagev1.CompanionInfo) *CompanionInfo {
	if pb == nil {
		return nil
	}
	return &CompanionInfo{
		FileName:   pb.GetFileName(),
		MIMEType:   pb.GetMimeType(),
		TotalBytes: pb.GetTotalBytes(),
		Kind:       companionKindFromProto(pb.GetKind()),
	}
}

func downloadChunkFromProto(pb *imessagev1.DownloadAttachmentResponse) (DownloadChunk, bool, error) {
	switch pb.WhichPayload() {
	case imessagev1.DownloadAttachmentResponse_Header_case:
		h := pb.GetHeader()
		return DownloadHeader{
			Info:      attachmentInfoFromProto(h.GetAttachment()),
			Companion: companionInfoFromProto(h.GetCompanion()),
		}, true, nil
	case imessagev1.DownloadAttachmentResponse_PrimaryChunk_case:
		return DownloadPrimaryChunk{Data: pb.GetPrimaryChunk()}, true, nil
	case imessagev1.DownloadAttachmentResponse_CompanionChunk_case:
		return DownloadCompanionChunk{Data: pb.GetCompanionChunk()}, true, nil
	default:
		return nil, false, nil
	}
}
