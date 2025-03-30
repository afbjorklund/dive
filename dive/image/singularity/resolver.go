package singularity

import (
	"fmt"
	"os"

	"github.com/wagoodman/dive/dive/image"
)

type fileResolver struct{}

func NewResolverFromFile() *fileResolver {
	return &fileResolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *fileResolver) Name() string {
	return "sif"
}

func (r *fileResolver) Fetch(path string) (*image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, err := NewSifFile(reader)
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}

func (r *fileResolver) Build(args []string) (*image.Image, error) {
	return nil, fmt.Errorf("build option not supported for singularity file resolver")
}

func (r *fileResolver) Extract(id string, l string, p string) error {
	return fmt.Errorf("not implemented")
}
