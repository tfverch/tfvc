package output

import (
	"fmt"
	"strings"

	goversion "github.com/hashicorp/go-version"
	"github.com/liamg/clinch/terminal"
	"github.com/liamg/tml"
)

const passed, passedInt = "PASSED", 0
const warning, warningInt = "WARNING", 1
const failed, failedInt = "FAILED", 2
const terraform = "terraform"

type Updates []Update

type Update struct {
	Type               string
	Path               string
	Name               string
	Source             string
	VersionConstraints goversion.Constraints
	Version            goversion.Version
	LatestMatching     goversion.Version
	LatestOverall      goversion.Version
	Status             string
	StatusInt          int
	Message            string
	Resolution         string
}

func (u *Update) DefaultOutput() {
	width, _ := terminal.Size()
	const detailsPad = 21
	if width <= 0 {
		width = 80
	}
	out := ""
	if u.Type == terraform {
		out += tml.Sprintf("\n<italic>%s</italic>", u.Type)
	} else {
		out += tml.Sprintf("\n<italic>%s '%s'</italic>", u.Type, u.Name)
	}
	switch u.Status {
	case passed:
		out += tml.Sprintf(" <bold><green>PASSED</green></bold>")
	case warning:
		out += tml.Sprintf(" <bold><yellow>WARNING</yellow></bold>")
	case failed:
		out += tml.Sprintf(" <bold><red>FAILED</red></bold>")
	}
	out += tml.Sprintf(" <bold>%s</bold>\n", u.Message)
	out += tml.Sprintf("<darkgrey>%s</darkgrey>\n\n", strings.Repeat("─", width))
	out += tml.Sprintf("  <dim>Resolution</dim>\n  <darkgrey>%s</darkgrey>\n  %s\n\n", strings.Repeat("─", len(u.Resolution)), u.Resolution)
	out += tml.Sprintf("  <dim>Details</dim>\n  <darkgrey>%s</darkgrey>\n", strings.Repeat("─", len(u.Source)+detailsPad))
	out += tml.Sprintf("  <dim>Type:</dim>                %s\n", u.Type)
	out += tml.Sprintf("  <dim>Path:</dim>                %s\n", u.Path)
	out += tml.Sprintf("  <dim>Name:</dim>                %s\n", u.Name)
	out += tml.Sprintf("  <dim>Source:</dim>              %s\n", u.Source)
	out += tml.Sprintf("  <dim>Version Constraints:</dim> %s\n", u.VersionConstraints.String())
	out += tml.Sprintf("  <dim>Version:</dim>             %s\n", u.Version.String())
	out += tml.Sprintf("  <dim>Latest Match:</dim>        %s\n", u.LatestMatching.String())
	out += tml.Sprintf("  <dim>Latest Overall:</dim>      %s\n", u.LatestOverall.String())
	out += tml.Sprintf("\n<darkgrey>%s</darkgrey>", strings.Repeat("─", width))
	tml.Printf("%s\n\n", out)
}

func (u *Update) SetUpdateStatus() { //nolint:gocognit
	vSegs, oSegs, mSegs := u.Version.Segments(), u.LatestOverall.Segments(), u.LatestMatching.Segments()
	u.Status, u.StatusInt = passed, passedInt
	u.Message = "No issues detected"
	u.Resolution = "No issues were detected with the current configuration"
	if u.Version.String() != "" && u.LatestOverall.String() != "" && u.LatestOverall.GreaterThan(&u.Version) {
		u.Status, u.StatusInt = warning, warningInt
		u.Message = "Configured version does not match the latest available version"
		u.Resolution = tml.Sprintf("Consider using the latest version of %s", thisOrNot(u.Type))
	}
	if u.Type == "provider" && u.Version.String() != "" && u.LatestMatching.GreaterThan(&u.Version) {
		u.Status, u.StatusInt = warning, warningInt
		u.Message = "Latest match newer than .terraform.lock.hcl config"
		u.Resolution = "Consider running 'terraform init -upgrade' to upgrade providers and modules to the latest matching versions"
	}
	if len(oSegs) > 0 && len(mSegs) > 0 && u.LatestOverall.GreaterThan(&u.LatestMatching) {
		u.Status, u.StatusInt = warning, warningInt
		u.Message = "Version constraint does not match the latest available version"
		u.Resolution = tml.Sprintf("Consider amending this version constraint to include the latest available version of %s", thisOrNot(u.Type))
	}
	if (len(oSegs) > 0 && len(mSegs) > 0) && (oSegs[0] > mSegs[0]) {
		u.Status, u.StatusInt = failed, failedInt
		u.Message = "Outdated major version"
		u.Resolution = tml.Sprintf("Consider migrating to the latest major version of %s", thisOrNot(u.Type))
	}
	if (len(oSegs) > 0 && len(vSegs) > 0) && (oSegs[0] > vSegs[0]) {
		u.Status, u.StatusInt = failed, failedInt
		u.Message = "Outdated major version"
		u.Resolution = tml.Sprintf("Consider migrating to the latest major version of %s", thisOrNot(u.Type))
	}
	if u.VersionConstraints == nil {
		u.Status, u.StatusInt = failed, failedInt
		u.Message = "Missing version constraints"
		u.Resolution = tml.Sprintf("Configure version constraints for %s", thisOrNot(u.Type))
	}
	if u.Type == "provider" && u.Source == "" {
		u.Status, u.StatusInt = failed, failedInt
		u.Message = "Missing provider definition"
		u.Resolution = tml.Sprintf("Configure source and version constraint for this provider in the `required_providers` block")
		u.LatestOverall = goversion.Version{}
	}
}

func thisOrNot(t string) string {
	if t == terraform {
		return terraform
	}
	return fmt.Sprintf("this %s", t)
}

// Need to assess the following as part of the readme update work

// func (u Updates) Format(w io.Writer, as Format) error {
// 	switch as {
// 	case FormatJSON:
// 		return u.WriteJSON(w)
// 	case FormatJSONL:
// 		return u.WriteJSONL(w)
// 	case FormatMarkdown:
// 		return u.WriteMarkdown(w)
// 	case FormatMarkdownWide:
// 		return u.WriteMarkdownWide(w)
// 	case FormatJUnit:
// 		return u.WriteJUnit(w)
// 	}
// 	return nil
// }

// func (u Updates) WriteJSONL(w io.Writer) error {
// 	enc := json.NewEncoder(w)
// 	enc.SetEscapeHTML(false)
// 	for _, item := range u {
// 		if err := enc.Encode(item); err != nil {
// 			return fmt.Errorf("encode json: %w", err)
// 		}
// 	}
// 	return nil
// }

// func (u Updates) WriteJSON(w io.Writer) error {
// 	enc := json.NewEncoder(w)
// 	enc.SetEscapeHTML(false)
// 	return enc.Encode(u)
// }

// func (u Updates) WriteMarkdownWide(w io.Writer) error {
// 	table := tablewriter.NewWriter(w)
// 	table.SetHeader([]string{"Type", "Name", "Path", "Source", "Constraint", "Version", "Latest matching", "Latest"})
// 	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
// 	table.SetCenterSeparator("|")
// 	rows := make([][]string, 0, len(u))
// 	for _, item := range u {
// 		// update := ""
// 		// switch {
// 		// case item.MatchingUpdate:
// 		// 	update = "Y"
// 		// case item.NonMatchingUpdate:
// 		// 	update = "(Y)"
// 		// case item.Version == "":
// 		// 	update = "?"
// 		// }
// 		row := []string{item.Type, item.Name, item.Path, item.Source, item.VersionConstraintString, item.VersionString, item.LatestMatching, item.LatestOverall}
// 		rows = append(rows, row)
// 	}
// 	table.AppendBulk(rows)
// 	table.Render()
// 	return nil
// }

// func (u Updates) WriteMarkdown(w io.Writer) error {
// 	table := tablewriter.NewWriter(w)
// 	table.SetHeader([]string{"Type", "Name", "Constraint", "Version", "Latest matching", "Latest"})
// 	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
// 	table.SetCenterSeparator("|")
// 	rows := make([][]string, 0, len(u))
// 	for _, item := range u {
// 		// update := ""
// 		// switch {
// 		// case item.MatchingUpdate:
// 		// 	update = "Y"
// 		// case item.NonMatchingUpdate:
// 		// 	update = "(Y)"
// 		// case item.Version == "":
// 		// 	update = "?"
// 		// }
// 		row := []string{item.Type, item.Name, item.VersionConstraintString, item.VersionString, item.LatestMatching, item.LatestOverall}
// 		rows = append(rows, row)
// 	}
// 	table.AppendBulk(rows)
// 	table.Render()
// 	return nil
// }

// func (u Updates) WriteJUnit(w io.Writer) error {
// 	testCases := make([]junit.JUnitTestCase, len(u))

// 	failures := 0
// 	for i, update := range u {
// 		testCase := junit.JUnitTestCase{
// 			Name:      update.Name,
// 			Classname: update.Path,
// 			Time:      "0",
// 		}
// 		success := !update.MatchingUpdate
// 		if !success {
// 			failures++
// 			testCase.Failure = &junit.JUnitFailure{
// 				Message:  fmt.Sprintf("Module version can be updated to %v (from %v)", update.LatestMatching, update.Version),
// 				Contents: "",
// 			}
// 		}
// 		testCases[i] = testCase
// 	}

// 	suites := junit.JUnitTestSuites{
// 		Suites: []junit.JUnitTestSuite{
// 			{
// 				Time:      "0",
// 				Tests:     len(u),
// 				Failures:  failures,
// 				TestCases: testCases,
// 			},
// 		},
// 	}

// 	if _, err := fmt.Fprint(w, xml.Header); err != nil {
// 		return fmt.Errorf("encode junit xml: %w", err)
// 	}
// 	enc := xml.NewEncoder(w)
// 	enc.Indent("", "  ")
// 	if err := enc.Encode(suites); err != nil {
// 		return fmt.Errorf("encode junit xml: %w", err)
// 	}
// 	if _, err := fmt.Fprintln(w); err != nil {
// 		return fmt.Errorf("encode junit xml: %w", err)
// 	}
// 	return nil
// }
