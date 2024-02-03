package generate

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/fanky5g/ponzu/internal/domain/services/contentgenerator"
	"github.com/spf13/cobra"
	"log"
)

func generateContentType(args []string, generators []interfaces.ContentGenerator) error {
	// parse type info from args
	gt, err := parseType(args)
	if err != nil {
		return fmt.Errorf("failed to parse type args: %s", err.Error())
	}

	for _, generator := range generators {
		for _, field := range gt.Fields {
			if err = generator.ValidateField(&field); err != nil {
				return err
			}
		}

		if err = generator.Generate(gt); err != nil {
			return err
		}
	}

	return nil
}

var generateCmd = &cobra.Command{
	Use:     "generate <generator type (,...fields)>",
	Aliases: []string{"gen", "g"},
	Short:   "generate boilerplate code for various Ponzu components",
	Long: `Generate boilerplate code for various Ponzu components, such as 'content'.

The command above will generate a file 'content/review.go' with boilerplate
methods, as well as struct definition, and corresponding field tags like:

type Review struct {
	Title  string   ` + "`json:" + `"title"` + "`" + `
	Body   string   ` + "`json:" + `"body"` + "`" + `
	Rating int      ` + "`json:" + `"rating"` + "`" + `
	Tags   []string ` + "`json:" + `"tags"` + "`" + `
}

The generate command will intelligently parse more sophisticated field names
such as 'field_name' and convert it to 'FieldName' and vice versa, only where
appropriate as per common Go idioms. Errors will be reported, but successful
generate commands return nothing.`,
	Example: `$ ponzu gen content review title:"string" body:"string" rating:"int" tags:"[]string"`,
}

var contentCmd = &cobra.Command{
	Use:     "content <namespace> <field> <field>...",
	Aliases: []string{"c"},
	Short:   "generates a new content type",
	RunE: func(cmd *cobra.Command, args []string) error {
		domainContentGenerator, err := contentgenerator.New()
		if err != nil {
			log.Fatalf("Failed to initialize domain content generator: %v", err)
		}

		return generateContentType(args, []interfaces.ContentGenerator{
			domainContentGenerator,
		})
	},
}

func RegisterCommandRecursive(parent *cobra.Command) {
	generateCmd.AddCommand(contentCmd)
	parent.AddCommand(generateCmd)
}
