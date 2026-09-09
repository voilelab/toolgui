package tgframe

import (
	"runtime/debug"
	"strings"
)

// modulePath is the module path go.mod declares. A dependent's build info
// lists this module under it.
const modulePath = "github.com/voilelab/toolgui"

// fallbackVersion is what Version reports when the build info carries no
// usable version for the module -- `go run`, `-buildvcs=false`, a build from a
// source tree with no VCS metadata, or one whose checkout has no tag to derive
// a version from.
//
// The Release workflow rewrites this line on the commit it tags, so a released
// tree always carries its own version. The value committed on dev is the last
// release it was bumped to.
const fallbackVersion = "v0.3.0"

// Version return the version of toolgui this binary was built against.
//
// It reads the module version out of the build info, so a released tag is
// reported without anything to maintain. A build off a tag reports the
// pseudo-version the toolchain derives from the commit. Only a build with no
// version at all falls back to the version recorded at release time.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallbackVersion
	}

	if info.Main.Path == modulePath && isRealVersion(info.Main.Version) {
		return info.Main.Version
	}

	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}

		// A replace directive points the module somewhere else; the
		// replacement carries the version that was actually built.
		if dep.Replace != nil {
			dep = dep.Replace
		}

		if isRealVersion(dep.Version) {
			return dep.Version
		}
	}

	return fallbackVersion
}

// isRealVersion reject what the toolchain reports when it has no version to
// go on: "", "(devel)", and the v0.0.0 pseudo-version it stamps from VCS when
// the checkout carries no tag to derive from -- a CI checkout fetched without
// tags, say. A pseudo-version off a real tag (v0.4.1-0.2026...-abc) does say
// which release the build sits after, so it is kept.
func isRealVersion(v string) bool {
	return v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0-")
}
