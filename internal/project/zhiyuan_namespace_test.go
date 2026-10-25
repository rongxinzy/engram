package project

import (
	"errors"
	"testing"
)

func TestResolveZhiYuanMemoryNamespaces(t *testing.T) {
	for _, name := range []string{
		"personal://zhiyuan-agent/user",
		"companion://zhiyuan-agent/00000000-0000-4000-8000-000000000001",
		"companion://zhiyuan-agent/pet-a",
	} {
		t.Run(name, func(t *testing.T) {
			checked := ""
			result, err := Resolve(ResolutionOptions{
				Mode: ResolutionCurrent, Explicit: name, RequireKnownExplicit: true,
				ProjectExists: func(project string) (bool, error) {
					checked = project
					return project == name, nil
				},
			})
			if err != nil || result.Project != name || checked != name || result.Source != SourceExplicitOverride {
				t.Fatalf("exact namespace resolution: result=%+v checked=%q error=%v", result, checked, err)
			}
		})
	}
}

func TestResolveZhiYuanMemoryNamespacesKeepValidation(t *testing.T) {
	for _, name := range []string{
		"/tmp/workspace", `C:\workspace`, "file:///tmp/workspace", "https://zhiyuan-agent/user",
		"personal://other/user", "personal://zhiyuan-agent/other", "personal://zhiyuan-agent/user/",
		"companion://other/pet-a", "companion://zhiyuan-agent/", "companion://zhiyuan-agent/../pet-a",
		"companion://zhiyuan-agent/pet-a/child", "companion://zhiyuan-agent/pet-a?all=true",
		"companion://zhiyuan-agent/pet-a#other", "companion://zhiyuan-agent/%2e%2e",
		"companion://zhiyuan-agent/pet\x00a", "companion://zhiyuan-agent/pet a",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Resolve(ResolutionOptions{Mode: ResolutionCurrent, Explicit: name})
			if !errors.Is(err, ErrInvalidProjectName) {
				t.Fatalf("invalid namespace %q: %v", name, err)
			}
		})
	}
	_, err := Resolve(ResolutionOptions{
		Mode: ResolutionCurrent, Explicit: "companion://zhiyuan-agent/unknown", RequireKnownExplicit: true,
		ProjectExists: func(string) (bool, error) { return false, nil },
	})
	var unknown *UnknownProjectError
	if !errors.As(err, &unknown) {
		t.Fatalf("unknown namespace must retain known-project validation: %v", err)
	}
}
