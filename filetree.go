package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Size      int64       `json:"size"`
	FileCount int64       `json:"file_count"`
	IsDir     bool        `json:"is_dir,omitempty"`
	Children  []*FileNode `json:"children"`
}

func (node *FileNode) String() string {
	return fmt.Sprintf("Name: %s, Size: %d", node.Name, node.Size)
}

func (node *FileNode) FindFiles(filename *regexp.Regexp) []*FileNode {

	if !node.IsDir {
		if filename.MatchString(node.Name) {
			return []*FileNode{node}
		}

		return nil
	}

	var results []*FileNode

	for _, child := range node.Children {
		// recursion
		results = append(results, child.FindFiles(filename)...)
	}

	return results
}

func (node *FileNode) GetDir(fullPath string) *FileNode {

	currentNode := node
	segments := strings.Split(fullPath, string(filepath.Separator))[1:]

Loop:
	for _, segment := range segments {

		for _, child := range currentNode.Children {

			if child.IsDir && child.Name == segment {
				currentNode = child
				continue Loop
			}
		}

		return nil
	}

	if currentNode.Path == fullPath {
		return currentNode
	}

	return nil
}

func New(path string) (*FileNode, error) {

	stats, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// base case, single file
	if !stats.IsDir() {
		return &FileNode{
			Name:      stats.Name(),
			Path:      path,
			Size:      stats.Size(),
			Children:  make([]*FileNode, 0),
			FileCount: 1,
		}, nil
	}

	// process directory
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	children := make([]*FileNode, 0)
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	count := int64(0)
	size := int64(0)

	for _, file := range files {

		// recursion
		node, err := New(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		children = append(children, node)

		// accumulate file count and size
		count += node.FileCount
		size += node.Size
	}

	return &FileNode{
		Name:      filepath.Base(path),
		Path:      path,
		IsDir:     true,
		Children:  children,
		FileCount: count,
		Size:      size,
	}, nil
}
