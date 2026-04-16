package cli

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	v1alpha1 "github.com/wiscotrashpanda/alloy/manifest/v1alpha1"
	"gopkg.in/yaml.v3"
)

// suffixAlphabet is the pool of characters used when disambiguating a manifest
// whose filename would otherwise collide with an existing file.
const suffixAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// runCreateGitHubRepository handles `smyth create-manifest github-repo`. It
// prompts for the minimal spec fields, validates the resulting manifest
// through alloy, and writes it to disk as YAML.
func runCreateGitHubRepository(args []string, stdin io.Reader, stdout io.Writer) error {
	s := newStyler(stdout)

	flags := flag.NewFlagSet("create-manifest github-repo", flag.ContinueOnError)
	flags.SetOutput(stdout)

	dir := flags.String("dir", ".", "Directory to write the manifest into")

	flags.Usage = func() {
		writeCreateGitHubRepositoryHelp(stdout, s)
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return err
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("create-manifest github-repo takes no positional arguments")
	}

	writeBanner(stdout, s)
	writeSectionHeader(stdout, s, "Authoring a GitHubRepository manifest")
	fmt.Fprintln(stdout, s.dim("  press enter to accept the default shown in brackets"))
	fmt.Fprintln(stdout)

	p := newPrompter(stdin, stdout)

	owner, err := p.askRequired("Repository owner (org or user)")
	if err != nil {
		return err
	}

	name, err := p.askRequired("Repository name")
	if err != nil {
		return err
	}

	name, err = disambiguateRepositoryName(*dir, owner, name, p)
	if err != nil {
		return err
	}

	visibility, err := p.askChoice("Visibility", []string{"public", "private", "internal"}, "private")
	if err != nil {
		return err
	}

	description, err := p.ask("Description (optional)", "")
	if err != nil {
		return err
	}

	homepage, err := p.ask("Homepage URL (optional)", "")
	if err != nil {
		return err
	}

	defaultBranch, err := p.ask("Default branch", "main")
	if err != nil {
		return err
	}

	autoInit, err := p.askBool("Auto-initialize with an initial commit", false)
	if err != nil {
		return err
	}

	topics, err := p.askList("Topics (comma separated, optional)")
	if err != nil {
		return err
	}

	metadataName := defaultMetadataName(owner, name)

	manifest := v1alpha1.NewGitHubRepositoryManifest(
		v1alpha1.Metadata{Name: metadataName},
		v1alpha1.GitHubRepositorySpec{
			Owner:         owner,
			Name:          name,
			Visibility:    visibility,
			Description:   description,
			Homepage:      homepage,
			DefaultBranch: defaultBranch,
			AutoInit:      autoInit,
			Topics:        topics,
		},
	)

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("generated manifest is invalid: %w", err)
	}

	encoded, err := encodeManifest(manifest)
	if err != nil {
		return err
	}

	outputPath, err := writeManifest(*dir, manifestFilename(owner, name), encoded)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout)
	writeSectionHeader(stdout, s, "Forged")
	fmt.Fprintf(stdout, "  %s wrote manifest to %s\n", s.green("✓"), s.bold(outputPath))
	fmt.Fprintf(stdout, "  %s review the file and commit it, then hand it to anvil to reconcile.\n", s.dim("›"))

	return nil
}

// writeSectionHeader renders a light horizontal rule so interactive output has
// some visual rhythm without feeling boxed-in.
func writeSectionHeader(w io.Writer, s *styler, title string) {
	fmt.Fprintf(w, "%s %s\n", s.forge("──"), s.bold(title))
}

// writeCreateGitHubRepositoryHelp emits the `--help` text for the subcommand.
// It keeps the banner off the help screen so scripts that run `--help` for
// discovery don't have to scroll past ASCII art.
func writeCreateGitHubRepositoryHelp(stdout io.Writer, s *styler) {
	fmt.Fprintf(stdout, `%s
  smyth create-manifest github-repo [--dir <path>]

Interactively prompts for the fields required to build a GitHubRepository
manifest. Nested specs (features, merge policy, branch protection, etc.) are
left out for now and can be added to the generated manifest by hand or by
follow-up commands.

%s
  %s %s
`,
		s.bold("Usage:"),
		s.bold("Flags:"),
		s.cyan("--dir <path>"),
		s.dim("Directory to write the manifest into (default: current directory)"),
	)
}

// disambiguateRepositoryName checks whether a manifest for owner/name already
// exists in dir. If it does, the user is warned and asked whether to continue.
// Declining aborts the command; confirming returns the repository name with a
// short random suffix appended so the new manifest does not clobber the
// existing one and the generated repo is unambiguously distinct.
func disambiguateRepositoryName(dir, owner, name string, p *prompter) (string, error) {
	path := filepath.Join(dir, manifestFilename(owner, name))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return name, nil
		}

		return "", fmt.Errorf("inspect %s: %w", path, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("expected manifest file but %s is a directory", path)
	}

	s := p.style
	fmt.Fprintf(p.writer, "\n%s a manifest for %s/%s already exists:\n", s.yellow("⚠"), s.bold(owner), s.bold(name))
	fmt.Fprintf(p.writer, "  %s\n", s.bold(path))
	fmt.Fprintln(p.writer, s.dim("  if this is the manifest you intended to update, edit it directly."))
	fmt.Fprintln(p.writer, s.dim("  continuing will append a random suffix so the new manifest describes"))
	fmt.Fprintln(p.writer, s.dim("  a different repository."))
	fmt.Fprintln(p.writer)

	cont, err := p.askBool("Continue anyway", false)
	if err != nil {
		return "", err
	}

	if !cont {
		return "", fmt.Errorf("aborted: manifest already exists at %s", path)
	}

	suffix, err := randomSuffix(4)
	if err != nil {
		return "", err
	}

	suffixed := fmt.Sprintf("%s-%s", name, suffix)
	fmt.Fprintf(p.writer, "  %s using repository name %s for the new manifest.\n", s.green("✓"), s.bold(suffixed))

	return suffixed, nil
}

// defaultMetadataName builds a conventional metadata.name from the owner and
// repository name. It keeps the value lowercase so it plays nicely with
// filesystem and DNS-style identifiers.
func defaultMetadataName(owner, name string) string {
	owner = strings.ToLower(strings.TrimSpace(owner))
	name = strings.ToLower(strings.TrimSpace(name))

	if owner == "" || name == "" {
		return name
	}

	return fmt.Sprintf("%s-%s", owner, name)
}

// manifestFilename returns the canonical filename for a GitHubRepository
// manifest belonging to owner/name.
func manifestFilename(owner, name string) string {
	owner = strings.ToLower(strings.TrimSpace(owner))
	name = strings.ToLower(strings.TrimSpace(name))

	return fmt.Sprintf("%s-%s.manifest.yaml", owner, name)
}

// randomSuffix returns a lowercase alphanumeric string of length n suitable
// for disambiguating a colliding manifest filename.
func randomSuffix(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate suffix: %w", err)
	}

	result := make([]byte, n)
	for i, b := range buf {
		result[i] = suffixAlphabet[int(b)%len(suffixAlphabet)]
	}

	return string(result), nil
}

// encodeManifest serializes a manifest to YAML with indentation that matches
// existing Anvil/Alloy examples.
func encodeManifest(manifest v1alpha1.GitHubRepositoryManifest) ([]byte, error) {
	var buf strings.Builder

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(manifest); err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("close encoder: %w", err)
	}

	return []byte(buf.String()), nil
}

// writeManifest writes the encoded manifest into dir under filename. It refuses
// to overwrite an existing file so that authoring a manifest never silently
// replaces prior work.
func writeManifest(dir, filename string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("ensure directory %q: %w", dir, err)
	}

	path := filepath.Join(dir, filename)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("refusing to overwrite existing file %s", path)
		}

		return "", fmt.Errorf("open %s: %w", path, err)
	}

	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}

	return path, nil
}
