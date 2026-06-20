import { createClient } from "@photon-ai/advanced-imessage";

const address = process.argv[2];
if (!address) {
  throw new Error("usage: node client.mjs host:port");
}

const CHAT = "any;-;alice@example.com";
const MESSAGE = "message-guid";
const REPLY = "reply-guid";
const ATTACHMENT = "attachment-guid";
const STICKER = "sticker-guid";
const PART = 2;

const EFFECT_CONFETTI = "com.apple.messages.effect.CKConfettiEffect";
const EFFECT_GENTLE = "com.apple.MobileSMS.expressivesend.gentle";
const FIXTURE_CREATED = "2026-06-20T12:00:00.123Z";
const FIXTURE_READ = "2026-06-20T12:01:00.123Z";
const FIXTURE_DELIVERED = "2026-06-20T12:02:00.123Z";
const FIXTURE_EDITED = "2026-06-20T12:03:00.123Z";
const FIXTURE_RETRACTED = "2026-06-20T12:04:00.123Z";
const FIXTURE_PLAYED = "2026-06-20T12:05:00.123Z";
const FIXTURE_EFFECT_PLAYED = "2026-06-20T12:06:00.123Z";

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function assertEqual(actual, expected, name) {
  if (actual !== expected) {
    throw new Error(`${name}: got ${JSON.stringify(actual)}, want ${JSON.stringify(expected)}`);
  }
}

function assertArray(actual, expected, name) {
  assertEqual(JSON.stringify(actual), JSON.stringify(expected), name);
}

function assertDate(actual, expected, name) {
  assert(actual instanceof Date, `${name}: got non-Date ${Object.prototype.toString.call(actual)}`);
  assertEqual(actual.toISOString(), expected, name);
}

function expectMessage(message) {
  assertEqual(message.guid, "fixture-message-guid", "message.guid");
  assertEqual(message.subject, "fixture subject", "message.subject");
  assertDate(message.dateCreated, FIXTURE_CREATED, "message.dateCreated");
  assertDate(message.dateRead, FIXTURE_READ, "message.dateRead");
  assertDate(message.dateDelivered, FIXTURE_DELIVERED, "message.dateDelivered");
  assertDate(message.dateEdited, FIXTURE_EDITED, "message.dateEdited");
  assertDate(message.dateRetracted, FIXTURE_RETRACTED, "message.dateRetracted");
  assertDate(message.datePlayed, FIXTURE_PLAYED, "message.datePlayed");
  assertDate(message.dateExpressiveSendPlayed, FIXTURE_EFFECT_PLAYED, "message.dateExpressiveSendPlayed");
  assertEqual(message.sender?.address, "+15550001111", "message.sender.address");
  assertEqual(message.sender?.service, "iMessage", "message.sender.service");
  assertEqual(message.sender?.country, "us", "message.sender.country");
  assertEqual(message.isFromMe, true, "message.isFromMe");
  assertEqual(message.isSent, true, "message.isSent");
  assertEqual(message.isDelivered, true, "message.isDelivered");
  assertEqual(message.isDeliveredQuietly, true, "message.isDeliveredQuietly");
  assertEqual(message.didNotifyRecipient, true, "message.didNotifyRecipient");
  assertEqual(message.sendErrorCode, 12, "message.sendErrorCode");
  assertEqual(message.isAudioMessage, true, "message.isAudioMessage");
  assertEqual(message.isAutoReply, true, "message.isAutoReply");
  assertEqual(message.isSystemMessage, true, "message.isSystemMessage");
  assertEqual(message.isForward, true, "message.isForward");
  assertEqual(message.isDelayed, true, "message.isDelayed");
  assertEqual(message.isSpam, true, "message.isSpam");
  assertEqual(message.dataDetectorResultsPresent, true, "message.dataDetectorResultsPresent");
  assertEqual(message.isArchived, true, "message.isArchived");
  assertEqual(message.isServiceMessage, true, "message.isServiceMessage");
  assertEqual(message.isCorrupt, true, "message.isCorrupt");
  assertEqual(message.isExpirable, true, "message.isExpirable");
  assertEqual(message.shareStatus, 1, "message.shareStatus");
  assertEqual(message.shareDirection, 2, "message.shareDirection");
  assertEqual(message.itemType, "chatAction", "message.itemType");
  assertEqual(message.groupTitle, "Fixture Group", "message.groupTitle");
  assertEqual(message.chatActionType, 44, "message.chatActionType");
  assertEqual(message.threadOriginatorGuid, "thread-originator-guid", "message.threadOriginatorGuid");
  assertEqual(message.threadOriginatorPart, "thread-originator-part", "message.threadOriginatorPart");
  assertEqual(message.replyTargetGuid, "reply-target-guid", "message.replyTargetGuid");
  assertEqual(message.destinationCallerId, "destination-caller-id", "message.destinationCallerId");
  assertEqual(message.partCount, 3, "message.partCount");
  assertEqual(message.cachedRoomNames, "room-a,room-b", "message.cachedRoomNames");
  assertEqual(message.reactionTargetGuid, "reaction-target-guid", "message.reactionTargetGuid");
  assertEqual(message.reactionTargetPartIndex, 2, "message.reactionTargetPartIndex");
  assertEqual(message.reaction?.kind, "emoji", "message.reaction.kind");
  assertEqual(message.reaction?.emoji, ":)", "message.reaction.emoji");
  assertEqual(message.reactionSelected, true, "message.reactionSelected");
  assertArray(message.chatGuids, [CHAT, "any;+;group-guid"], "message.chatGuids");

  assertEqual(message.content.text, "fixture body", "message.content.text");
  assertEqual(message.content.balloonBundleId, "com.example.balloon", "message.content.balloonBundleId");
  assertEqual(message.content.expressiveSendStyleId, EFFECT_CONFETTI, "message.content.expressiveSendStyleId");
  assertEqual(message.content.attachments[0]?.guid, "fixture-attachment-guid", "message.content.attachments[0].guid");
  assertEqual(message.content.attachments[0]?.originalGuid, "original-attachment-guid", "message.content.attachments[0].originalGuid");
  assertEqual(message.content.attachments[0]?.fileName, "photo.jpg", "message.content.attachments[0].fileName");
  assertEqual(message.content.attachments[0]?.mimeType, "image/jpeg", "message.content.attachments[0].mimeType");
  assertEqual(message.content.attachments[0]?.uti, "public.jpeg", "message.content.attachments[0].uti");
  assertEqual(message.content.attachments[0]?.totalBytes, 123456, "message.content.attachments[0].totalBytes");
  assertEqual(message.content.attachments[0]?.isOutgoing, true, "message.content.attachments[0].isOutgoing");
  assertEqual(message.content.attachments[0]?.transferState, "finished", "message.content.attachments[0].transferState");
  assertEqual(message.content.attachments[0]?.isHidden, true, "message.content.attachments[0].isHidden");
  assertEqual(message.content.attachments[0]?.isSticker, true, "message.content.attachments[0].isSticker");
  assertEqual(message.content.attachments[0]?.companionKind, "live-photo-video", "message.content.attachments[0].companionKind");
  assertEqual(message.content.formatting[0]?.type, "bold", "message.content.formatting[0].type");
  assertEqual(message.content.formatting[1]?.effectName, "bloom", "message.content.formatting[1].effectName");
  assertEqual(message.content.mentions[0]?.address, "alice@example.com", "message.content.mentions[0].address");

  assertEqual(message.appliedReactions[0]?.messageGuid, "fixture-message-guid", "message.appliedReactions[0].messageGuid");
  assertEqual(message.appliedReactions[0]?.targetPartIndex, 1, "message.appliedReactions[0].targetPartIndex");
  assertEqual(message.appliedReactions[0]?.reaction.kind, "love", "message.appliedReactions[0].reaction.kind");
  assertEqual(message.appliedReactions[0]?.sender?.address, "+15550002222", "message.appliedReactions[0].sender.address");
  assertDate(message.appliedReactions[0]?.dateCreated, "2026-06-20T12:07:00.123Z", "message.appliedReactions[0].dateCreated");

  assertEqual(message.placedStickers[0]?.messageGuid, "fixture-message-guid", "message.placedStickers[0].messageGuid");
  assertEqual(message.placedStickers[0]?.targetPartIndex, 2, "message.placedStickers[0].targetPartIndex");
  assertEqual(message.placedStickers[0]?.sticker?.guid, "fixture-sticker-guid", "message.placedStickers[0].sticker.guid");
  assertEqual(message.placedStickers[0]?.placement?.x, 0.4, "message.placedStickers[0].placement.x");
  assertEqual(message.placedStickers[0]?.placement?.y, 0.6, "message.placedStickers[0].placement.y");
  assertEqual(message.placedStickers[0]?.placement?.scale, 1.25, "message.placedStickers[0].placement.scale");
  assertEqual(message.placedStickers[0]?.placement?.rotation, 0.5, "message.placedStickers[0].placement.rotation");
  assertEqual(message.placedStickers[0]?.placement?.width, 0.3, "message.placedStickers[0].placement.width");
}

function expectPage(page) {
  assertEqual(page.nextPageToken, "next-page-token", "page.nextPageToken");
  assertEqual(page.messages.length, 1, "page.messages.length");
  expectMessage(page.messages[0]);
}

async function collect(stream) {
  const out = [];
  for await (const event of stream) {
    out.push(event);
  }
  return out;
}

function eventName(event) {
  if (event.type === "group.changed") {
    return `${event.type}.${event.change.type}`;
  }
  if (event.type === "poll.changed") {
    return `${event.type}.${event.delta.type}`;
  }
  return event.type;
}

function expectEventNames(events, expected, name) {
  assertArray(events.map(eventName), expected, name);
}

const im = createClient({
  address,
  token: "interop-token",
  tls: false,
});

try {
  const replyTo = { guid: REPLY, partIndex: PART };
  const formatting = [
    { type: "bold", start: 0, length: 5 },
    { type: "effect", start: 6, length: 5, effect: "bloom" },
  ];
  const sendOptions = {
    replyTo,
    subject: "interop subject",
    effect: EFFECT_CONFETTI,
    enableDataDetection: true,
    enableLinkPreview: true,
    formatting,
    clientMessageId: "cmid-send-text",
  };

  expectMessage(await im.messages.sendText(CHAT, "hello from interop", sendOptions));
  expectMessage(await im.messages.sendAttachment(CHAT, ATTACHMENT, {
    replyTo,
    effect: EFFECT_GENTLE,
    isAudioMessage: true,
    clientMessageId: "cmid-send-attachment",
  }));
  expectMessage(await im.messages.sendMultipart(CHAT, [
    { text: "hello ", bubbleIndex: 0, formatting },
    { text: "@Alice", mentionedAddress: "alice@example.com", bubbleIndex: 1 },
    { attachmentGuid: ATTACHMENT, attachmentName: "photo.jpg", bubbleIndex: 2 },
  ], {
    replyTo,
    subject: "multipart subject",
    effect: EFFECT_CONFETTI,
    enableDataDetection: true,
    clientMessageId: "cmid-send-multipart",
  }));
  expectMessage(await im.messages.sendCustomizedMiniApp(CHAT, {
    appName: "Interop App",
    appStoreId: 1234567890123,
    extensionBundleId: "com.example.MessagesExtension",
    layout: {
      caption: "Caption",
      subcaption: "Subcaption",
      trailingCaption: "Trailing Caption",
      trailingSubcaption: "Trailing Subcaption",
      imageTitle: "Image Title",
      imageSubtitle: "Image Subtitle",
      summary: "Summary",
      image: new Uint8Array([1, 2, 3, 4]),
    },
    teamId: "TEAMID1234",
    url: "https://example.com/interop",
  }, { clientMessageId: "cmid-mini-app" }));
  expectMessage(await im.messages.edit(CHAT, MESSAGE, "edited text", {
    backwardCompatText: "edited fallback",
    partIndex: PART,
    clientMessageId: "cmid-edit",
  }));
  await im.messages.unsend(CHAT, MESSAGE, {
    partIndex: PART,
    clientMessageId: "cmid-unsend",
  });

  for (const [index, reaction] of [
    { kind: "love" },
    { kind: "like" },
    { kind: "dislike" },
    { kind: "laugh" },
    { kind: "emphasize" },
    { kind: "question" },
    { kind: "emoji", emoji: ":)" },
  ].entries()) {
    expectMessage(await im.messages.setReaction(CHAT, MESSAGE, reaction, index % 2 === 0, {
      partIndex: PART,
      clientMessageId: `cmid-reaction-${reaction.kind}`,
    }));
  }

  expectMessage(await im.messages.placeSticker(CHAT, MESSAGE, STICKER, {
    x: 0.4,
    y: 0.6,
    scale: 1.25,
    rotation: 0.5,
    width: 0.3,
  }, {
    partIndex: PART,
    clientMessageId: "cmid-sticker",
  }));
  await im.messages.notifySilenced(CHAT, MESSAGE, { clientMessageId: "cmid-notify" });

  const getMessage = im.messages.get.length >= 2
    ? await im.messages.get(CHAT, MESSAGE)
    : await im.messages.get(MESSAGE);
  expectMessage(getMessage);

  expectPage(await im.messages.listRecent({
    pageSize: 25,
    pageToken: "recent-page-token",
    isFromMe: true,
    isRead: false,
    before: new Date("2026-06-20T12:30:00.123Z"),
    after: new Date("2026-06-20T11:30:00.123Z"),
  }));
  expectPage(await im.messages.listInChat(CHAT, {
    pageSize: 10,
    pageToken: "chat-page-token",
    isFromMe: false,
    isRead: true,
    before: new Date("2026-06-20T12:45:00.123Z"),
    after: new Date("2026-06-20T11:45:00.123Z"),
  }));

  const media = await im.messages.getEmbeddedMedia(CHAT, MESSAGE);
  assertArray(Array.from(media.data), [9, 8, 7, 6], "embeddedMedia.data");
  assertEqual(media.mimeType, "image/png", "embeddedMedia.mimeType");

  const messageEvents = await collect(im.messages.subscribeEvents({ chat: CHAT }));
  expectEventNames(messageEvents, [
    "message.received",
    "message.edited",
    "message.read",
    "message.unsent",
    "message.reactionAdded",
    "message.reactionRemoved",
    "message.stickerPlaced",
  ], "message event names");
  expectMessage(messageEvents[0].message);

  const catchUpEvents = await collect(im.events.catchUp(42));
  expectEventNames(catchUpEvents, [
    "message.received",
    "message.edited",
    "message.read",
    "message.unsent",
    "message.reactionAdded",
    "message.reactionRemoved",
    "message.stickerPlaced",
    "chat.backgroundChanged",
    "chat.backgroundRemoved",
    "chat.markedRead",
    "chat.archived",
    "chat.unarchived",
    "group.changed.displayNameChanged",
    "group.changed.participantAdded",
    "group.changed.participantRemoved",
    "group.changed.participantLeft",
    "group.changed.iconChanged",
    "group.changed.iconRemoved",
    "poll.changed.created",
    "poll.changed.optionAdded",
    "poll.changed.voted",
    "poll.changed.unvoted",
    "catchup.complete",
  ], "catch-up event names");
  assertEqual(catchUpEvents.at(-1)?.headSequence, 3000, "catch-up headSequence");

  console.log(JSON.stringify({
    messageEvents: messageEvents.length,
    catchUpEvents: catchUpEvents.length,
  }));
} finally {
  await im.close();
}
