package utils

import "regexp"

// clusterIdentifierRe constrains a cluster identifier to the characters shared by every form
// OCM accepts for a cluster lookup: the internal ID (alphanumeric), the external ID (an RFC4122
// UUID, so hexadecimal plus dashes) and the cluster display name.
//
// This deliberately validates the character class rather than the shape of either format, and
// mainly serves as a guard against injections.
var clusterIdentifierRe = regexp.MustCompile(`^[a-zA-Z0-9-]{1,64}$`)

// IsValidClusterIdentifier reports whether the identifier is safe to interpolate into an OCM
// search expression. Identifiers taken from untrusted sources (alert payloads, for example) must
// pass this check before being used in a lookup.
func IsValidClusterIdentifier(identifier string) bool {
	return clusterIdentifierRe.MatchString(identifier)
}
