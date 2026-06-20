package imessage

// TransferState is the delivery state of an attachment's bytes.
type TransferState string

const (
	// TransferStatePending means transfer has not started.
	TransferStatePending TransferState = "pending"
	// TransferStateTransferring means transfer is in progress.
	TransferStateTransferring TransferState = "transferring"
	// TransferStateFailed means transfer failed.
	TransferStateFailed TransferState = "failed"
	// TransferStateFinished means the bytes are available.
	TransferStateFinished TransferState = "finished"
	// TransferStateUnavailable means the attachment bytes are unavailable.
	TransferStateUnavailable TransferState = "unavailable"
	// TransferStateUnknown is an unrecognized state.
	TransferStateUnknown TransferState = "unknown"
)

// CompanionKind classifies the sidecar that ships with an attachment.
type CompanionKind string

const (
	// CompanionLivePhotoVideo is the video sidecar of a Live Photo.
	CompanionLivePhotoVideo CompanionKind = "live-photo-video"
	// CompanionUnknown is an unrecognized or absent companion kind.
	CompanionUnknown CompanionKind = "unknown"
)

// AttachmentInfo is metadata about a message attachment.
type AttachmentInfo struct {
	// GUID is the attachment identifier.
	GUID string
	// OriginalGUID, if set, is the attachment this one superseded via an edit.
	OriginalGUID string
	// FileName is the original file name.
	FileName string
	// MIMEType is the MIME type.
	MIMEType string
	// UTI is the Apple Uniform Type Identifier (e.g. "public.jpeg").
	UTI string
	// TotalBytes is the size in bytes.
	TotalBytes int64
	// IsOutgoing reports whether the local user sent the attachment.
	IsOutgoing bool
	// TransferState is the delivery state of the bytes.
	TransferState TransferState
	// IsHidden reports whether the attachment is hidden from the conversation.
	IsHidden bool
	// IsSticker reports whether the attachment is a sticker.
	IsSticker bool
	// CompanionKind is the sidecar kind, or "" if there is none.
	CompanionKind CompanionKind
}

// CompanionInfo is metadata about an attachment's sidecar.
type CompanionInfo struct {
	FileName   string
	MIMEType   string
	TotalBytes int64
	Kind       CompanionKind
}

// AttachmentInput is the payload for uploading an attachment, optionally with a
// Live Photo companion.
type AttachmentInput struct {
	// FileName is the file name to record.
	FileName string
	// Data is the primary file's bytes.
	Data []byte
	// Companion, if non-nil, is the sidecar's bytes (e.g. a Live Photo video).
	Companion *CompanionInput
}

// CompanionInput is the sidecar payload for [AttachmentInput].
type CompanionInput struct {
	Data []byte
	Kind CompanionKind
}

// UploadAttachmentResult is the outcome of an upload.
type UploadAttachmentResult struct {
	Attachment AttachmentInfo
	Companion  *CompanionInfo
}

// DownloadChunk is one frame of a streamed attachment download. Use a type
// switch over [DownloadHeader], [DownloadPrimaryChunk], and
// [DownloadCompanionChunk].
type DownloadChunk interface{ isDownloadChunk() }

// DownloadHeader is the first frame of a download, carrying metadata.
type DownloadHeader struct {
	Info      AttachmentInfo
	Companion *CompanionInfo
}

// DownloadPrimaryChunk carries a slice of the primary file's bytes.
type DownloadPrimaryChunk struct{ Data []byte }

// DownloadCompanionChunk carries a slice of the companion sidecar's bytes.
type DownloadCompanionChunk struct{ Data []byte }

func (DownloadHeader) isDownloadChunk()         {}
func (DownloadPrimaryChunk) isDownloadChunk()   {}
func (DownloadCompanionChunk) isDownloadChunk() {}
