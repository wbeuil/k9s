package view

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

func showPodsWithLabels(app *App, path string, sel map[string]string) {
	log.Debug().Msgf("SHOWING POD FOR %#v", sel)
	var labels []string
	for k, v := range sel {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	showPods(app, path, strings.Join(labels, ","), "")
}

func showPods(app *App, path, labelSel, fieldSel string) {
	log.Debug().Msgf("SHOW PODS %q -- %q -- %q", path, labelSel, fieldSel)
	app.switchNS("")

	v := NewPod(client.GVR("v1/pods"))
	v.SetContextFn(podCtx(path, labelSel, fieldSel))
	v.GetTable().SetColorerFn(render.Pod{}.ColorerFunc())

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	if err := app.inject(v); err != nil {
		app.Flash().Err(err)
	}
}

func podCtx(path, labelSel, fieldSel string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		ctx = context.WithValue(ctx, internal.KeyLabels, labelSel)
		return context.WithValue(ctx, internal.KeyFields, fieldSel)
	}
}

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		return nil, errors.New("No application found in context")
	}

	return app, nil
}

// AsKey maps a string representation of a key to a tcell key.
func asKey(key string) (tcell.Key, error) {
	for k, v := range tcell.KeyNames {
		if v == key {
			return k, nil
		}
	}

	return 0, fmt.Errorf("No matching key found %s", key)
}

// FwFQN returns a fully qualified ns/name:container id.
func fwFQN(po, co string) string {
	return po + ":" + co
}

func isTCPPort(p string) bool {
	return !strings.Contains(p, "UDP")
}

// ContainerID computes container ID based on ns/po/co.
func containerID(path, co string) string {
	ns, n := client.Namespaced(path)
	po := strings.Split(n, "-")[0]

	return ns + "/" + po + ":" + co
}

// UrlFor computes fq url for a given benchmark configuration.
func urlFor(cfg config.BenchConfig, port string) string {
	host := "localhost"
	if cfg.HTTP.Host != "" {
		host = cfg.HTTP.Host
	}

	path := "/"
	if cfg.HTTP.Path != "" {
		path = cfg.HTTP.Path
	}

	return "http://" + host + ":" + port + path
}

func fqn(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}