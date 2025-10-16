package transpiler

// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-inputs-param
type GalaxyTypeValidator string

const (
	GalaxyTypeValidatorText           GalaxyTypeValidator = "text"
	GalaxyTypeValidatorInteger        GalaxyTypeValidator = "integer"
	GalaxyTypeValidatorFloat          GalaxyTypeValidator = "float"
	GalaxyTypeValidatorBoolean        GalaxyTypeValidator = "boolean"
	GalaxyTypeValidatorGenomeBuild    GalaxyTypeValidator = "genomebuild"
	GalaxyTypeValidatorSelect         GalaxyTypeValidator = "select"
	GalaxyTypeValidatorColor          GalaxyTypeValidator = "color"
	GalaxyTypeValidatorDataColumn     GalaxyTypeValidator = "data_column"
	GalaxyTypeValidatorHidden         GalaxyTypeValidator = "hidden"
	GalaxyTypeValidatorHiddenData     GalaxyTypeValidator = "hidden_data"
	GalaxyTypeValidatorBaseURL        GalaxyTypeValidator = "baseurl"
	GalaxyTypeValidatorFile           GalaxyTypeValidator = "file"
	GalaxyTypeValidatorFTPFile        GalaxyTypeValidator = "ftpfile"
	GalaxyTypeValidatorData           GalaxyTypeValidator = "data"
	GalaxyTypeValidatorDataCollection GalaxyTypeValidator = "data_collection"
	GalaxyTypeValidatorDrillDown      GalaxyTypeValidator = "drill_down"
)

var galaxyTypeValidators = [...]GalaxyTypeValidator{
	GalaxyTypeValidatorText,
	GalaxyTypeValidatorInteger,
	GalaxyTypeValidatorFloat,
	GalaxyTypeValidatorBoolean,
	GalaxyTypeValidatorGenomeBuild,
	GalaxyTypeValidatorSelect,
	GalaxyTypeValidatorColor,
	GalaxyTypeValidatorDataColumn,
	GalaxyTypeValidatorHidden,
	GalaxyTypeValidatorHiddenData,
	GalaxyTypeValidatorBaseURL,
	GalaxyTypeValidatorFile,
	GalaxyTypeValidatorFTPFile,
	GalaxyTypeValidatorData,
	GalaxyTypeValidatorDataCollection,
	GalaxyTypeValidatorDrillDown,
}
