package singularity

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"

	"github.com/CalebQ42/squashfs"
	"github.com/sylabs/sif/v2/pkg/sif"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

type SifFile struct {
	digest string
	tree   *filetree.FileTree
}

func NewSifFile(sifFile *os.File) (*SifFile, error) {
	img := &SifFile{}

	image, err := sif.LoadContainer(sifFile)
	if err != nil {
		return nil, err
	}

	desc, err := image.GetDescriptor(sif.WithDataType(sif.DataPartition))
	if err != nil {
		return nil, err
	}
	offset := desc.Offset()

	data, err := desc.GetData()
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)

	img.digest = "sha256:" + hex.EncodeToString(digest[:])

	squashfsReader, err := squashfs.NewReaderAtOffset(sifFile, offset)
	if err != nil {
		return nil, err
	}
	tree, err := processFilesystem("FS", squashfsReader)
	if err != nil {
		return nil, err
	}

	img.tree = tree

	return img, nil
}

func processFilesystem(name string, reader *squashfs.Reader) (*filetree.FileTree, error) {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos, err := getFileList(reader)
	if err != nil {
		return nil, err
	}

	for _, element := range fileInfos {
		if element.Path == "." {
			continue
		}
		tree.FileSize += uint64(element.Size)

		_, _, err := tree.AddPath(element.Path, element)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func getFileList(squashfsReader *squashfs.Reader) ([]filetree.FileInfo, error) {
	var files []filetree.FileInfo

	err := fs.WalkDir(squashfsReader, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		files = append(files, filetree.NewFileInfoFromDirEntry(squashfsReader, d, path))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (img *SifFile) ToImage() (*image.Image, error) {
	layer := &image.Layer{
		Id:      "squashfs",
		Index:   0,
		Command: "",
		Size:    img.tree.FileSize,
		Tree:    img.tree,
		Names:   []string{"(unavailable)"},
		Digest:  img.digest,
	}

	return &image.Image{
		Trees:  []*filetree.FileTree{img.tree},
		Layers: []*image.Layer{layer},
	}, nil
}
