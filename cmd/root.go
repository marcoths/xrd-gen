package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/marcoths/xgen/internal/xrd"
	"github.com/spf13/cobra"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"os/exec"
	"path/filepath"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xgen",
	Short: "Generate Crossplane XRDs from Go structs",
	Long: `xrd-gen is a command-line interface (CLI) to generate Crossplane Composite Resource Definitions (XRDs) from
Go structs.`,
	Example: ` # Generate XRDs for all CRDs in the examples/deploy directory.
xrd-gen --path=examples/apis --output-dir=examples/deploy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString("path")
		cobra.CheckErr(err)
		out, err := cmd.Flags().GetString("output-dir")
		cobra.CheckErr(err)

		if err := checkFlag(path); err != nil {
			return fmt.Errorf("failed to parse path: %w", err)
		}
		if err := checkFlag(out); err != nil {
			return fmt.Errorf("failed to parse output-dir: %w", err)
		}

		if err := ensureDeps(); err != nil {
			return err
		}

		return run(path, out)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("path", "p", "", "path to the Go structs directory")
	rootCmd.Flags().StringP("output-dir", "o", "", "path to the output directory")
}

func run(path, out string) error {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = apiextv1.AddToScheme(sch)
	//if err := callDirectly(path, out); err != nil {
	//	return fmt.Errorf("failed to call directly: %w", err)
	//}
	if err := generateCRDs(path, out); err != nil {
		return fmt.Errorf("failed to generate CRDs: %w", err)
	}
	decoder := serializer.NewCodecFactory(sch).UniversalDeserializer()
	dirContent, err := FilePathWalkDir(out)
	if err != nil {
		return fmt.Errorf("failed to find dirContent: %w", err)
	}
	for _, f := range dirContent {
		content, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("failed to read file: %v\n", err)
			continue
		}
		obj, gvk, err := decoder.Decode(content, nil, nil)
		if err != nil {
			fmt.Printf("failed to decode file: %v\n", err)
			continue
		}
		if gvk.Kind != "CustomResourceDefinition" {
			fmt.Printf("skipping non crd: %s\n", gvk.Kind)
			continue
		}
		crd := obj.(*apiextv1.CustomResourceDefinition)

		gen := &xrd.Generator{
			Group:          crd.Spec.Group,
			Version:        crd.Spec.Versions[0].Name,
			Kind:           crd.Spec.Names.Kind,
			Plural:         crd.Spec.Names.Plural,
			OverrideFields: []xrd.OverrideField{},
			ManifestsPath:  out,
		}
		crdJSON, _ := json.Marshal(crd)
		if err := gen.Generate(crdJSON); err != nil {
			return fmt.Errorf("failed to make xrd: %w", err)
		}
	}
	return nil

}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func ensureDeps() error {
	if !commandExists("controller-gen") {
		color.Red("controller-gen not found. Please install by running `make controller-gen` in the root of the project")
		return errors.New("there was a problem checking dependencies")
	}
	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func generateCRDs(path, out string) error {
	schemapatch := fmt.Sprintf("schemapatch:manifests=./%s", out)
	output := fmt.Sprintf("output:dir=%s", out)
	paths := fmt.Sprintf("paths=./%s/...", path)
	cmd := exec.Command("controller-gen", "crd", schemapatch, output, paths)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkFlag(flag string) error {
	if flag == "" {
		return errors.New("argument must be provided")
	}
	d, err := isDir(flag)
	if err != nil {
		return err
	}
	if !d {
		return errors.New(flag + " argument must be a directory")
	}
	return nil
}

func isDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open path: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.IsDir(), nil
}
