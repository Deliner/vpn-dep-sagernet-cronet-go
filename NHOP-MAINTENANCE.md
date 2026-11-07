# Maintained Go wrapper source

This branch starts at SagerNet/cronet-go commit
`095cb9e060468c80ca31d977c020dca82d101a8f` (Chromium 154 generation).
The Go wrapper implementation and its module requirements are unchanged.
Upstream history and licenses remain available through the parent commit.

The `lib/*` native binary modules and `all` aggregation module are excluded from
this branch. They are independent Go modules, not part of the root wrapper
module. Keeping every platform's prebuilt binaries in the selected Git tree
caused the public Go proxy's initial fork fetch to time out. This separation
keeps the wrapper source fetch small and makes native artifact provenance an
explicit, separate build input.

The NaiveProxy Git submodule still selects exactly
`7c64cae24cd7bae18d9c603fdddc73b623c04462`; its URL now names our source fork.
No native binary is rebuilt or certified by this source-only commit. Existing
upstream native modules remain on their original commits, and native ABI,
payload acceptance and C/C++ advisory applicability require their own checks.

This change does not claim to fix the intermittent MASQUE UDP observation.
