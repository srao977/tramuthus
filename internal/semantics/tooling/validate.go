package tooling

import (
	"os"

	"fin_feedsat_1/internal/semantics"
)

func ValidatePath(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = semantics.Parse(raw)
	return err
}

func ValidateCatalog() error {
	return semantics.Validate(Catalog())
}
