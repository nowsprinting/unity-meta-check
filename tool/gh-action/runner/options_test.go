package runner

import (
	"github.com/DeNA/unity-meta-check/options"
	"github.com/DeNA/unity-meta-check/resultfilter"
	"github.com/DeNA/unity-meta-check/tool/gh-action/inputs"
	"github.com/DeNA/unity-meta-check/tool/unity-meta-autofix/autofix"
	"github.com/DeNA/unity-meta-check/tool/unity-meta-check-github-pr-comment/github"
	"github.com/DeNA/unity-meta-check/tool/unity-meta-check-github-pr-comment/l10n"
	"github.com/DeNA/unity-meta-check/unity/checker"
	"github.com/DeNA/unity-meta-check/util/globs"
	"github.com/DeNA/unity-meta-check/util/typedpath"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"reflect"
	"testing"
)

func TestNewValidateFunc(t *testing.T) {
	githubComAPIEndpoint, err := url.Parse("https://api.github.com")
	if err != nil {
		t.Errorf("want nil, got %#v", err)
		return
	}

	tmpl := &l10n.Template{
		SuccessMessage: "s",
		FailureMessage: "f",
		StatusHeader:   "s",
		FilePathHeader: "f",
		StatusMissing:  "s",
		StatusDangling: "s",
	}

	cases := map[string]struct {
		RootDirRel         typedpath.RawPath
		Cwd                typedpath.RawPath
		Inputs             inputs.Inputs
		Token              github.Token
		DetectedTargetType checker.TargetType
		BuiltIgnoredGlobs  []globs.Glob
		ReadPayload        *inputs.PullRequestEventPayload
		ReadTmpl           *l10n.Template
		Expected           *Options
	}{
		"all default": {
			Inputs: inputs.Inputs{
				LogLevel:   "INFO",
				TargetType: "auto-detect",
				TargetPath: typedpath.NewRootRawPath("path", "to", "project"),
			},
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{"ignore*"},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{"ignore*"},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"ignore-case": {
			Inputs: inputs.Inputs{
				LogLevel:   "INFO",
				TargetPath: typedpath.NewRootRawPath("path", "to", "project"),
				TargetType: "auto-detect",
				IgnoreCase: true,
			},
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                true,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{},
					IgnoreCase:   true,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"explicit unity-project": {
			Inputs: inputs.Inputs{
				LogLevel:   "INFO",
				TargetPath: typedpath.NewRootRawPath("path", "to", "project"),
				TargetType: "unity-project",
			},
			BuiltIgnoredGlobs: []globs.Glob{"ignore*"},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					TargetType:                checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{"ignore*"},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"explicit unity-project-sub-dir": {
			Cwd: typedpath.NewRootRawPath("path", "to", "project"),
			Inputs: inputs.Inputs{
				LogLevel:   "INFO",
				TargetPath: typedpath.NewRawPath("Assets", "Foo"),
				TargetType: "unity-project-sub-dir",
			},
			BuiltIgnoredGlobs: []globs.Glob{"ignore*"},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project", "Assets", "Foo"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					TargetType:                checker.TargetTypeIsUnityProjectSubDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{"ignore*"},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"explicit upm-package": {
			Cwd: typedpath.NewRootRawPath("path", "to", "project"),
			Inputs: inputs.Inputs{
				LogLevel:   "INFO",
				TargetType: "upm-package",
				TargetPath: typedpath.NewRawPath("Packages", "com.example.pkg"),
			},
			BuiltIgnoredGlobs: []globs.Glob{"ignore*"},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project", "Packages", "com.example.pkg"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					TargetType:                checker.TargetTypeIsUnityProjectSubDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{"ignore*"},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"enable junit": {
			Inputs: inputs.Inputs{
				LogLevel:     "INFO",
				TargetPath:   typedpath.NewRootRawPath("path", "to", "project"),
				TargetType:   "auto-detect",
				EnableJUnit:  true,
				JUnitXMLPath: typedpath.NewRawPath("junit.xml"),
			},
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{},
					IgnoreCase:   false,
				},
				EnableJUnit:     true,
				JUnitOutPath:    typedpath.NewRawPath("junit.xml"),
				EnablePRComment: false,
				EnableAutofix:   false,
			},
		},
		"enable pr-comment with lang": {
			Inputs: inputs.Inputs{
				LogLevel:             "INFO",
				TargetPath:           typedpath.NewRootRawPath("path", "to", "project"),
				TargetType:           "auto-detect",
				EnablePRComment:      true,
				PRCommentLang:        "ja",
				PRCommentEventPath:   typedpath.NewRootRawPath("github", "workflows", "event.json"),
				PRCommentSendSuccess: true,
			},
			Token:              "T0K3N",
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{},
			ReadPayload: &inputs.PullRequestEventPayload{
				PullRequest: inputs.PullRequest{Number: 2},
				Repository: inputs.Repository{
					URL:   "https://api.github.com/repos/Codertocat/Hello-World",
					Name:  "Hello-World",
					Owner: inputs.User{Login: "Codertocat"},
				},
			},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: true,
				PRCommentOpts: &github.Options{
					Tmpl:          &l10n.Ja,
					SendIfSuccess: true,
					Token:         "T0K3N",
					APIEndpoint:   githubComAPIEndpoint,
					Owner:         "Codertocat",
					Repo:          "Hello-World",
					PullNumber:    2,
				},
			},
		},
		"enable pr-comment with a template file": {
			Inputs: inputs.Inputs{
				LogLevel:              "INFO",
				TargetPath:            typedpath.NewRootRawPath("path", "to", "project"),
				TargetType:            "auto-detect",
				EnablePRComment:       true,
				PRCommentTmplFilePath: typedpath.NewRawPath("tmpl.json"),
				PRCommentEventPath:    typedpath.NewRootRawPath("github", "workflows", "event.json"),
				PRCommentSendSuccess:  true,
			},
			Token:              "T0K3N",
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{},
			ReadTmpl:           tmpl,
			ReadPayload: &inputs.PullRequestEventPayload{
				PullRequest: inputs.PullRequest{Number: 2},
				Repository: inputs.Repository{
					URL:   "https://api.github.com/repos/Codertocat/Hello-World",
					Name:  "Hello-World",
					Owner: inputs.User{Login: "Codertocat"},
				},
			},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: true,
				PRCommentOpts: &github.Options{
					Tmpl:          tmpl,
					SendIfSuccess: true,
					Token:         "T0K3N",
					APIEndpoint:   githubComAPIEndpoint,
					Owner:         "Codertocat",
					Repo:          "Hello-World",
					PullNumber:    2,
				},
			},
		},
		"enable autofix": {
			RootDirRel: ".",
			Inputs: inputs.Inputs{
				LogLevel:      "INFO",
				TargetPath:    typedpath.NewRootRawPath("path", "to", "project"),
				TargetType:    "auto-detect",
				EnableAutofix: true,
				AutofixGlobs:  []string{"Assets/Allow/To/Fix/*"},
			},
			DetectedTargetType: checker.TargetTypeIsUnityProjectRootDirectory,
			BuiltIgnoredGlobs:  []globs.Glob{},
			Expected: &Options{
				RootDirAbs: typedpath.NewRootRawPath("path", "to", "project"),
				CheckerOpts: &checker.Options{
					IgnoreCase:                false,
					IgnoreSubmodulesAndNested: false,
					// NOTE: same as the value of DetectedTargetType.
					TargetType: checker.TargetTypeIsUnityProjectRootDirectory,
				},
				FilterOpts: &resultfilter.Options{
					IgnoreDangling: false,
					// NOTE: same as the value of BuiltIgnoredGlobs.
					IgnoredGlobs: []globs.Glob{},
					IgnoreCase:   false,
				},
				EnableJUnit:     false,
				EnablePRComment: false,
				EnableAutofix:   true,
				AutofixOpts: &autofix.Options{
					RootDirAbs:   typedpath.NewRootRawPath("path", "to", "project"),
					RootDirRel:   ".",
					AllowedGlobs: []globs.Glob{"Assets/Allow/To/Fix/*"},
				},
			},
		},
	}

	for desc, c := range cases {
		t.Run(desc, func(t *testing.T) {
			validate := NewValidateFunc(
				options.FakeRootDirValidator(c.Cwd),
				options.StubUnityProjectDetector(c.DetectedTargetType, nil),
				options.StubIgnoredPathBuilder(c.BuiltIgnoredGlobs, nil),
				autofix.StubOptionsBuilderWithRootDirAbsAndRel(c.RootDirRel),
				l10n.StubTemplateFileReader(c.ReadTmpl, nil),
				inputs.StubReadEventPayload(c.ReadPayload, nil),
			)

			opts, err := validate(c.Inputs, c.Token)
			if err != nil {
				t.Errorf("want nil, got %#v", err)
				return
			}

			if !reflect.DeepEqual(opts, c.Expected) {
				t.Error(cmp.Diff(c.Expected, opts))
			}
		})
	}
}
