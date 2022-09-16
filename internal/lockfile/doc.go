// This is a bastardised version of the internal depsfile and addr packages from the terraform
// source code i.e. "github.com/hashicorp/terraform/internal/addrs" and
// "github.com/hashicorp/terraform/internal/depsfile". We simply need to load the .terraform.lock.hcl
// file and decode the version and version constraint for each provider. So this lockfile package
// contains only the essential functions etc. from the two internal modules listed above.
// Yes, this is not nice but it works.
package lockfile
