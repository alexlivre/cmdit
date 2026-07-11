package command

import "testing"

func TestFilter_Empty(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "test", Label: "Test Action"})
	results := r.Filter("")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "file.save", Label: "Save File"})
	r.Register(Action{ID: "edit.copy", Label: "Copy"})

	results := r.Filter("SAVE")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'SAVE', got %d", len(results))
	}
	if results[0].ID != "file.save" {
		t.Errorf("expected file.save, got %s", results[0].ID)
	}
}

func TestFilter_Unicode(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "test", Label: "Ação Português"})
	results := r.Filter("ação")
	if len(results) != 1 {
		t.Errorf("expected 1 result for unicode query, got %d", len(results))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "test", Label: "Save"})
	results := r.Filter("xyz")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "test", Label: "Test", Shortcut: "Ctrl+T"})
	all := r.All()
	if len(all) != 1 {
		t.Errorf("expected 1 action, got %d", len(all))
	}
	if all[0].ID != "test" {
		t.Errorf("expected ID 'test', got '%s'", all[0].ID)
	}
}

func TestRegistry_Filter_ByLabel(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "file.save", Label: "Save File"})
	r.Register(Action{ID: "edit.copy", Label: "Copy"})
	results := r.Filter("save")
	if len(results) != 1 || results[0].ID != "file.save" {
		t.Errorf("expected file.save, got %v", results)
	}
}

func TestRegistry_Filter_ByID(t *testing.T) {
	r := NewRegistry()
	r.Register(Action{ID: "file.save", Label: "Save File"})
	results := r.Filter("file")
	if len(results) != 1 {
		t.Errorf("expected 1 result for ID filter, got %d", len(results))
	}
}
