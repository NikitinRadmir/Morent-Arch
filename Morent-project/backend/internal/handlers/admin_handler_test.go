package handlers

import "testing"

func TestImageExtensionUsesURLWhenSupported(t *testing.T) {
	t.Parallel()

	got := imageExtension("https://cdn.example.com/cars/kodiaq.webp?auto=format", "image/jpeg")
	if got != ".webp" {
		t.Fatalf("expected .webp, got %q", got)
	}
}

func TestImageExtensionFallsBackToContentType(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
		"image/avif": ".avif",
		"text/html":  ".jpg",
	}

	for contentType, want := range cases {
		got := imageExtension("https://cdn.example.com/image", contentType)
		if got != want {
			t.Fatalf("content type %q: expected %q, got %q", contentType, want, got)
		}
	}
}

func TestIsAllowedRemoteImageType(t *testing.T) {
	t.Parallel()

	allowed := []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/avif; charset=utf-8"}
	for _, contentType := range allowed {
		if !isAllowedRemoteImageType(contentType) {
			t.Fatalf("expected %q to be allowed", contentType)
		}
	}

	rejected := []string{"text/html", "image/svg+xml", "application/json", ""}
	for _, contentType := range rejected {
		if isAllowedRemoteImageType(contentType) {
			t.Fatalf("expected %q to be rejected", contentType)
		}
	}
}

func TestSanitizeObjectPart(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"BMW X5":       "bmw_x5",
		"  Mercedes! ": "mercedes",
		"":             "unknown",
		"***":          "unknown",
		"Skoda--Rapid": "skoda_rapid",
	}

	for input, want := range cases {
		if got := sanitizeObjectPart(input); got != want {
			t.Fatalf("sanitizeObjectPart(%q): expected %q, got %q", input, want, got)
		}
	}
}
