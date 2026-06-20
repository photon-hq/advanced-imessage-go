package imessage

import "testing"

func TestParseChatGUID(t *testing.T) {
	valid := []string{
		"iMessage;-;+15551234567",
		"iMessage;+;chat1234567890",
		"SMS;-;person@example.com",
	}
	for _, v := range valid {
		if _, err := ParseChatGUID(v); err != nil {
			t.Errorf("ParseChatGUID(%q) unexpected error: %v", v, err)
		}
	}

	invalid := []string{
		"",
		"+15551234567",          // bare address
		"iMessage;;identifier",  // empty kind
		"iMessage;-;",           // empty identifier
		";-;identifier",         // empty service
		"iMessage;x;identifier", // bad kind
		"iMessage:-:identifier", // wrong separator
	}
	for _, v := range invalid {
		if _, err := ParseChatGUID(v); err == nil {
			t.Errorf("ParseChatGUID(%q) expected error, got nil", v)
		}
	}
}

func TestChatGUIDIsGroup(t *testing.T) {
	group, err := ParseChatGUID("iMessage;+;chatabc")
	if err != nil {
		t.Fatal(err)
	}
	if !group.IsGroup() {
		t.Error("expected IsGroup() == true for a + GUID")
	}

	direct, err := ParseChatGUID("iMessage;-;+15551234567")
	if err != nil {
		t.Fatal(err)
	}
	if direct.IsGroup() {
		t.Error("expected IsGroup() == false for a - GUID")
	}
}
