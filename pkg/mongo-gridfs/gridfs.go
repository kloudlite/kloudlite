package mongogridfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type gfs struct {
	bucket *gridfs.Bucket
}

type GridFs interface {
	Upload(ctx context.Context, filename, source string) error
	Download(ctx context.Context, filename, destination string) error
	Upsert(ctx context.Context, filename, source string) error
	DeleteById(id string) error
	GetAllFiles() ([]GridfsFile, error)
	FetchFileRef(ctx context.Context, filename string) (*GridfsFile, error)
	DeleteAllWithFilename(filename string) error
}

// Delete implements GridFs
func (g *gfs) DeleteById(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return g.bucket.Delete(_id)
}

func (g *gfs) DeleteAllWithFilename(filename string) error {
	gf, err := g.GetAllFiles()
	if err != nil {
		return err
	}
	for _, gf2 := range gf {
		if gf2.Name == filename {
			if err := g.DeleteById(gf2.Id); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete implements GridFs
func (g *gfs) Upsert(ctx context.Context, filename, source string) error {
	if err := g.DeleteAllWithFilename(filename); err != nil {
		return err
	}

	return g.Upload(ctx, filename, source)
}

// Download implements GridFs
func (g *gfs) Download(ctx context.Context, filename, destination string) error {
	gf, err := g.FetchFileRef(ctx, filename)
	if err != nil {
		return err
	}

	id, err := primitive.ObjectIDFromHex(gf.Id)
	if err != nil {
		return err
	}
	fileBuffer := bytes.NewBuffer(nil)
	if _, err := g.bucket.DownloadToStream(id, fileBuffer); err != nil {
		panic(err)
	}

	if err := os.WriteFile(destination, fileBuffer.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

type Filter map[string]interface{}

type GridfsFile struct {
	Id     string `bson:"_id"`
	Name   string `bson:"filename"`
	Length int64  `bson:"length"`
}

func (g *gfs) GetAllFiles() ([]GridfsFile, error) {
	filter := bson.D{{}}
	// filter := bson.D{{"length", bson.D{{"$lt", 1500}}}}
	cursor, err := g.bucket.Find(filter)
	if err != nil {
		return nil, err
	}
	var foundFiles []GridfsFile
	if err = cursor.All(context.TODO(), &foundFiles); err != nil {
		return nil, err
	}

	return foundFiles, nil
}

// Search implements GridFs
func (g *gfs) FetchFileRef(ctx context.Context, filename string) (*GridfsFile, error) {

	cursor, err := g.bucket.Find(Filter{"filename": filename})
	if err != nil {
		return nil, err
	}

	var foundFiles []GridfsFile

	if err = cursor.All(context.TODO(), &foundFiles); err != nil {
		return nil, err
	}

	if len(foundFiles) == 0 {
		return nil, nil
	}

	return &foundFiles[0], nil

}

// Upload implements GridFs
func (g *gfs) Upload(ctx context.Context, filename, source string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}

	uploadOpts := options.GridFSUpload()
	objectID, err := g.bucket.UploadFromStream(filename, io.Reader(file), uploadOpts)
	if err != nil {
		return err
	}
	fmt.Printf("file %s uploaded with ID %s", filename, objectID)
	return nil
}
