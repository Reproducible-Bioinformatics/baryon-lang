package importer

import (
	"encoding/xml"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/galaxy"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/transpiler"
)

type GalaxyImporter struct {
	galaxyTool *galaxy.Tool
	transpiler.TranspilerBase
}

var _ Importer = (*GalaxyImporter)(nil)

// Export implements Importer.
func (g *GalaxyImporter) Export() (string, error) {
	g.Buffer.Reset()

	g.WriteLine("(bala %s (", g.galaxyTool.Name)
	g.SetIndentLevel(g.GetIndentLevel() + 1)
	g.WriteLine("; Parameter definition")

	// Parameters
	for _, param := range g.galaxyTool.Inputs.Param {
		if param.Type != "enum" {
			g.WriteLine("(%s %s (desc \"%s\"))",
				param.Name,
				param.Type,
				param.Help)
		} else {
			g.WriteLine("(%s (enum ( ", param.Name)
			g.SetIndentLevel(g.GetIndentLevel() + 1)
			for _, option := range param.Options {
				g.WriteLine("\"%s\"", option.Value)
			}
			g.SetIndentLevel(g.GetIndentLevel() - 1)
			g.WriteLine(") (desc \"%s\"))", param.Help)
		}
	}
	g.WriteLine("", "")

	// run_docker implementation.
	g.WriteLine("; Implementation: run_docker")
	g.WriteLine("(run_docker", "")
	g.SetIndentLevel(g.GetIndentLevel() + 1)
	g.WriteLine("(image \"%s\")", g.galaxyTool.Requirements.Container[0].Value)
	g.WriteLine("(arguments \"%s\")", g.galaxyTool.Command.Value)
	g.WriteLine(")", "")
	g.WriteLine("", "")

	// Outputs
	g.WriteLine("(outputs")
	for _, output := range g.galaxyTool.Outputs.Data {
		g.WriteLine("(%s %s %s)", output.Name, output.Format, output.Label)
	}
	g.WriteLine(")", "")
	g.WriteLine("", "")

	g.WriteLine("(desc", "")
	g.SetIndentLevel(g.GetIndentLevel() + 1)
	g.WriteLine("\"%s\"", g.galaxyTool.Description)
	g.SetIndentLevel(g.GetIndentLevel() - 1)
	g.WriteLine(")", "")
	g.WriteLine("", "")

	g.SetIndentLevel(g.GetIndentLevel() - 1)
	g.WriteLine(")", "")

	return "", nil
}

// Import implements Importer.
func (g *GalaxyImporter) Import(content []byte) error {
	g.galaxyTool = &galaxy.Tool{}
	err := xml.Unmarshal(content, g.galaxyTool)
	if err != nil {
		return err
	}
	return nil
}
