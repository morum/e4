package ssh

import "testing"

func TestAutoCompleteRoomIDCompletesSingleMatch(t *testing.T) {
	completion := autoCompleteRoomID("join AB", len("join AB"), []string{"ABC123", "QWE999"})
	if !completion.ok {
		t.Fatal("expected autocomplete to succeed")
	}
	if completion.showMatches {
		t.Fatal("expected direct completion, not match listing")
	}
	if completion.line != "join ABC123 " {
		t.Fatalf("expected full room ID completion, got %q", completion.line)
	}
	if completion.pos != len(completion.line) {
		t.Fatalf("expected cursor to move to end of completed line, got %d", completion.pos)
	}
}

func TestAutoCompleteRoomIDExtendsSharedPrefix(t *testing.T) {
	completion := autoCompleteRoomID("watch AB", len("watch AB"), []string{"ABEF34", "ABEZ99", "ZZZZ99"})
	if !completion.ok {
		t.Fatal("expected autocomplete to extend shared prefix")
	}
	if completion.showMatches {
		t.Fatal("expected shared-prefix completion, not match listing")
	}
	if completion.line != "watch ABE" {
		t.Fatalf("expected shared prefix completion, got %q", completion.line)
	}
	if completion.pos != len(completion.line) {
		t.Fatalf("expected cursor to move with completion, got %d", completion.pos)
	}
}

func TestAutoCompleteRoomIDRequiresJoinOrWatchCommand(t *testing.T) {
	if completion := autoCompleteRoomID("list AB", len("list AB"), []string{"ABC123"}); completion.ok {
		t.Fatal("expected autocomplete to ignore unrelated commands")
	}
}

func TestAutoCompleteRoomIDReturnsMatchListRequest(t *testing.T) {
	completion := autoCompleteRoomID("join AB", len("join AB"), []string{"ABCD12", "ABEF34"})
	if !completion.ok {
		t.Fatal("expected autocomplete to report ambiguous matches")
	}
	if !completion.showMatches {
		t.Fatal("expected autocomplete to request match listing")
	}
	if completion.query != "join AB" {
		t.Fatalf("expected autocomplete query to be tracked, got %q", completion.query)
	}
	if len(completion.matches) != 2 {
		t.Fatalf("expected both matching room IDs, got %#v", completion.matches)
	}
}
