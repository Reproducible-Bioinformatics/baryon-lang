// Original source for this file: https://github.com/Reproducible-Bioinformatics/baryon/blob/main/tool/tool.go
package galaxy

import (
	"encoding/xml"
	"fmt"
)

// Validable represents a validable object.
type Validable interface{ Validate() error }

// Tool provides a representation of a Galaxy Tool xml file schema.
//
// You can find the current schema here:
// https://docs.galaxyproject.org/en/master/dev/schema.html
type Tool struct {
	XMLName xml.Name `xml:"tool"`
	// The value is displayed in the tool menu immediately following the hyperlink
	// for the tool (based on the name attribute of the <tool> tag set described
	// above).
	//
	// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-description
	Description    string          `xml:"description"`
	EdamTopics     *EdamTopics     `xml:"edam_topics,omitempty"`
	EdamOperations *EdamOperations `xml:"edam_operations,omitempty"`
	Xrefs          *Xrefs          `xml:"xrefs,omitempty"`
	Creator        *Creator        `xml:"creator,omitempty"`
	Requirements   *Requirements   `xml:"requirements"`
	Command        *Command        `xml:"command"`
	Inputs         *Inputs         `xml:"inputs"`
	Outputs        *Outputs        `xml:"outputs"`
	Id             string          `xml:"id,attr"`
	Name           string          `xml:"name,attr"`
}

// Container tag set for the <edam_topic> tags. A tool can have any number of
// EDAM topic references.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-edam-topics
type EdamTopics struct {
	XMLName   xml.Name    `xml:"edam_topics,omitempty"`
	EdamTopic []EdamTopic `xml:"edam_topic,omitempty"`
}

type EdamTopic string

// Container tag set for the <edam_operation> tags. A tool can have any number
// of EDAM operation references.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-edam-operations
type EdamOperations struct {
	XMLName       xml.Name        `xml:"edam_operations"`
	EdamOperation []EdamOperation `xml:"edam_operation"`
}

type EdamOperation string

// Container tag set for the <xref> tags. A tool can refer multiple reference
// IDs.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-xrefs
type Xrefs struct {
	XMLName xml.Name `xml:"xrefs"`
	Xref    []Xref   `xml:"xref"`
}

// The xref element specifies reference information according to a catalog.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-xrefs-xref
type Xref struct {
	XMLName xml.Name `xml:"xref"`
	// Type of reference - currently bio.tools, bioconductor, and biii
	// are the only supported options.
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// The creator(s) of this work. See schema.org/creator.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-creator
type Creator struct {
	XMLName      xml.Name      `xml:"creator,omitempty"`
	Person       []Person      `xml:"person,omitempty"`
	Organization *Organization `xml:"organization,omitempty"`
}

// Describes a person. Tries to stay close to schema.org/Person.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-creator-person
type Person struct {
	XMLName xml.Name `xml:"person,omitempty"`
	Name    string   `xml:"name,omitempty"`
}

// Describes an organization. Tries to stay close to schema.org/Organization.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-creator-organization
type Organization struct {
	XMLName xml.Name `xml:"organization,omitempty"`
	Name    string   `xml:"name,omitempty"`
}

// This is a container tag set for the requirement, resource and container tags
// described in greater detail below. requirements describe software packages
// and other individual computing requirements required to execute a tool,
// while containers describe Docker or Singularity containers that should be
// able to serve as complete descriptions of the runtime of a tool.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-requirements
type Requirements struct {
	XMLName     xml.Name      `xml:"requirements"`
	Requirement []Requirement `xml:"requirement,omitempty"`
	Container   []Container   `xml:"container,omitempty"`
}

// This tag set is contained within the <requirements> tag set. Third party
// programs or modules that the tool depends upon are included in this tag set.
//
// When a tool runs, Galaxy attempts to resolve these requirements (also called
// dependencies). requirements are meant to be abstract and resolvable by
// multiple different dependency resolvers (e.g. conda, the Galaxy Tool Shed
// dependency management system, or environment modules).
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-requirements-requirement
type Requirement struct {
	XMLName xml.Name `xml:"requirement"`
	Type    string   `xml:"type,attr"`
	Version string   `xml:"version,attr"`
}

// This tag set is contained within the ‘requirements’ tag set. Galaxy can be
// configured to run tools within Docker or Singularity containers - this tag
// allows the tool to suggest possible valid containers for this tool.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-requirements-container
type Container struct {
	XMLName xml.Name `xml:"container"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",chardata"`
	Volumes []VolumeMapping
}

// Implements Validable.
func (c Container) Validate() error {
	var allowedType map[string]struct{} = map[string]struct{}{
		"docker":      {},
		"singularity": {},
	}
	if _, ok := allowedType[c.Type]; !ok {
		return fmt.Errorf("Type \"%s\" is not an allowed type.", c.Type)
	}
	return nil
}

// This tag specifies how Galaxy should invoke the tool’s executable, passing
// its required input parameter values (the command line specification links
// the parameters supplied in the form with the actual tool executable).
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-command
type Command struct {
	XMLName xml.Name `xml:"command"`
	Value   string   `xml:",cdata"`
}

// Consists of all elements that define the tool’s input parameters.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-inputs
type Inputs struct {
	XMLName xml.Name `xml:"inputs"`
	Param   []Param  `xml:"param"`
}

// Contained within the <inputs> tag set - each of these specifies a field that
// will be displayed on the tool form. Ultimately, the values of these form
// fields will be passed as the command line parameters to the tool’s
// executable.
//
// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-inputs-param
type Param struct {
	XMLName         xml.Name `xml:"param"`
	Type            string   `xml:"type,attr"`
	Name            string   `xml:"name,omitempty,attr"`
	Value           string   `xml:"value,omitempty,attr"`
	Options         []Option `xml:"option"`
	Argument        string   `xml:"argument,omitempty"`
	Label           string   `xml:"label,omitempty"`
	Help            string   `xml:"help,omitempty"`
	Optional        bool     `xml:"optional,omitempty"`
	RefreshOnChange bool     `xml:"refresh_on_change,omitempty"`
}

// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-inputs-param-option
type Option struct {
	XMLName       xml.Name `xml:"option"`
	Value         string   `xml:"value,attr"`
	CanonicalName string   `xml:",innerxml"`
}

// Implements Validable.
func (p Param) Validate() error {
	var allowedType map[string]struct{} = map[string]struct{}{
		"text":            {},
		"integer":         {},
		"float":           {},
		"boolean":         {},
		"genomebuild":     {},
		"select":          {},
		"color":           {},
		"data_column":     {},
		"hidden":          {},
		"hidden_data":     {},
		"baseurl":         {},
		"file":            {},
		"ftpfile":         {},
		"data":            {},
		"data_collection": {},
		"drill_down":      {},
	}
	if _, ok := allowedType[p.Type]; !ok {
		return fmt.Errorf("Type \"%s\" is not an allowed type.", p.Type)
	}
	if p.Optional && p.Value == "" {
		return fmt.Errorf("Non optional parameter has no value specified.")
	}
	return nil
}

// Container tag set for the <data> and <collection> tag sets. The files and
// collections created by tools as a result of their execution are named by
// Galaxy. You specify the number and type of your output files using the
// contained <data> and <collection> tags. These may be passed to your tool
// executable through using line variables just like the parameters described
// in the <inputs> documentation.
//
// https://docs.galaxyproject.org/en/master/dev/schema.html#tool-outputs
type Outputs struct {
	XMLName xml.Name `xml:"outputs"`
	Data    []Data
}

// This tag set is contained within the <outputs> tag set, and it defines the
// output data description for the files resulting from the tool’s execution.
// The value of the attribute label can be acquired from input parameters or
// metadata in the same way that the command line parameters are (discussed in
// the <command> tag set section above).
//
// https://docs.galaxyproject.org/en/master/dev/schema.html#tool-outputs-data
type Data struct {
	XMLName xml.Name `xml:"data"`
	Format  string   `xml:"format,omitempty,attr"`
	Name    string   `xml:"name,omitempty,attr"`
	Label   string   `xml:"label,omitempty,attr"`
}

// Implements Validable.
func (d Data) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("Name has no value specified.")
	}
	if d.Format == "" {
		return fmt.Errorf("Format has no value specified.")
	}
	return nil
}

// TODO: Integrate this with galaxy
//   - research tool volume mapping.
type VolumeMapping struct {
	HostPath  string
	GuestPath string
}
