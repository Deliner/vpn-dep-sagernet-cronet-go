# Maintained Cronet source

This branch starts at SagerNet/cronet-go commit
`095cb9e060468c80ca31d977c020dca82d101a8f` (Chromium 154 generation).
The Go wrapper runtime implementation and its module requirements are unchanged.
Upstream history and licenses remain available through the parent commit.

The `lib/*` native binary modules and `all` aggregation module are excluded from
this branch. They are independent Go modules, not part of the root wrapper
module. Keeping every platform's prebuilt binaries in the selected Git tree
caused the public Go proxy's initial fork fetch to time out. This separation
keeps the wrapper source fetch small and makes native artifact provenance an
explicit, separate build input.

The NaiveProxy Git submodule selects our source fork at
`f1e434fc78b9b1d9a04c17571984a0b6ab836d98`, descended from the original
`7c64cae24cd7bae18d9c603fdddc73b623c04462` snapshot. It contains six Chromium
production backports for range validation, proxy socket/ConnectJob lifetime,
proxy authentication endpoint selection, proxy-resolution request lifetime,
and certificate consistency before connection-based HTTP authentication.
Each native commit identifies the upstream fix and the limits of its tests.

The Go range regression exercises full and split reads, cacheable/no-store
responses, and positive controls through the actual native URLRequest API.
The baseline returned bytes past the advertised range; the first four
backports passed this regression and Naive H2/H3 TCP/UDP on AMD64 and ARM64.
That evidence predates the last two defensive backports and does not certify
new library bytes. Native builds, ABI/payload acceptance and the remaining
C/C++ advisory review must run again before updating release pins.
No native binary is rebuilt or published by this source-only commit.

This change does not claim to fix the intermittent MASQUE UDP observation.

The selected native source also backports zlib THM-ZLIB-001 to both the
generic and Chromium SIMD decoders. The bounded source regression checks
invalid-header/reset/dictionary handling and fixed-window controls under
ASan/UBSan. New native libraries still require Linux AMD64/ARM64 rebuild
and range/Naive H2/H3 acceptance before publication.
