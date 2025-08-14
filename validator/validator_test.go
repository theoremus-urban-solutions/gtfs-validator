package validator

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

// dummyValidator is a no-op validator used to verify the interface wiring
type dummyValidator struct{ called *bool }

func (d dummyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config Config) {
	if d.called != nil {
		*d.called = true
	}
	container.AddNotice(notice.NewBaseNotice("dummy_notice", notice.INFO, map[string]interface{}{"ok": true}))
}

func TestValidatorInterface_Wiring(t *testing.T) {
	// Minimal feed
	tmpLoader, err := parser.LoadFromDirectory(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}
	t.Cleanup(func() {
		if err := tmpLoader.Close(); err != nil {
			t.Errorf("Failed to close loader: %v", err)
		}
	})

	container := notice.NewNoticeContainer()
	called := false
	v := dummyValidator{called: &called}
	v.Validate(tmpLoader, container, Config{})

	if !called {
		t.Fatalf("expected validator to be called")
	}
	if len(container.GetNotices()) == 0 {
		t.Fatalf("expected a notice to be added by dummy validator")
	}
}
