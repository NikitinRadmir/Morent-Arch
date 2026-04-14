package server

import (
	"net/http"
	"strings"
)

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

type RouteGroup struct {
	Prefix string
	Routes []Route
}

// RegisterRouteGroups регистрирует группы роутов на DefaultServeMux.
func RegisterRouteGroups(groups []RouteGroup, wrapper func(http.HandlerFunc) http.HandlerFunc) {
	RegisterRouteGroupsToMux(nil, groups, wrapper)
}

// RegisterRouteGroupsToMux регистрирует группы роутов на переданном mux.
// Поддерживаются:
//   - статические пути (как раньше, с семантикой net/http по слэшам в конце)
//   - "динамические" сегменты вида "/:id" внутри Path.
func RegisterRouteGroupsToMux(mux *http.ServeMux, groups []RouteGroup, wrapper func(http.HandlerFunc) http.HandlerFunc) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	for _, group := range groups {
		if group.Prefix == "" {
			group.Prefix = "/"
		}

		handler := buildGroupHandler(group)
		if wrapper != nil {
			handler = wrapper(handler)
		}

		base := group.Prefix
		// Регистрируем и сам префикс, и префикс со слэшем,
		// чтобы обрабатывать и "/cars", и "/cars/", и "/cars/1".
		mux.HandleFunc(base, handler)
		if !strings.HasSuffix(base, "/") {
			mux.HandleFunc(base+"/", handler)
		}
	}
}

func buildGroupHandler(group RouteGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		pathMatched := false

		for _, route := range group.Routes {
			if !matchPath(group.Prefix, route.Path, path) {
				continue
			}
			pathMatched = true

			// Семантика метода такая же, как раньше:
			// пустой Method = любой, плюс всегда пропускаем OPTIONS.
			if route.Method != "" && method != route.Method && method != http.MethodOptions {
				continue
			}

			route.Handler(w, r)
			return
		}

		if pathMatched {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.NotFound(w, r)
	}
}

// matchPath реализует старое поведение для статических путей и добавляет
// поддержку сегментов вида ":id" ("/:id", "/car/:id" и т.п.).
func matchPath(prefix, routePath, requestPath string) bool {
	if !strings.HasPrefix(requestPath, prefix) {
		return false
	}

	// Динамические сегменты вида "/:id"
	if strings.Contains(routePath, ":") {
		rest := strings.TrimPrefix(requestPath, prefix)
		rest = strings.Trim(rest, "/")
		pat := strings.Trim(routePath, "/")

		if pat == "" && rest == "" {
			return true
		}

		var restParts []string
		if rest != "" {
			restParts = strings.Split(rest, "/")
		}
		var patParts []string
		if pat != "" {
			patParts = strings.Split(pat, "/")
		}

		if len(restParts) != len(patParts) {
			return false
		}

		for i := range patParts {
			pp := patParts[i]
			rp := restParts[i]
			if strings.HasPrefix(pp, ":") {
				// Любой непустой сегмент.
				if rp == "" {
					return false
				}
				continue
			}
			if pp != rp {
				return false
			}
		}
		return true
	}

	// Статические пути (совместимо с прежней логикой).
	// Путь группы "/" или пустой routePath интерпретируем как корень группы.
	if routePath == "" || routePath == "/" {
		if requestPath == prefix {
			return true
		}
		// "/cars/" тоже считаем корнем.
		if !strings.HasSuffix(prefix, "/") && requestPath == prefix+"/" {
			return true
		}
		return false
	}

	full := prefix + routePath

	if requestPath == full {
		return true
	}
	if strings.HasSuffix(full, "/") && strings.HasPrefix(requestPath, full) {
		return true
	}

	return false
}

