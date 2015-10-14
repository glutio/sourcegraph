package storage

import (
	"fmt"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type file struct {
	fs     *fileSystem
	name   *sourcegraph.StorageName
	offset int64
}

func (f *file) Name() string {
	return f.name.Name
}

func (f *file) String() string {
	return fmt.Sprintf("File(%q, FileSystem=%v)", f.name.Name, f.fs)
}

func (f *file) Read(p []byte) (n int, err error) {
	resp, grpcErr := f.fs.client.Storage.Read(f.fs.ctx, &sourcegraph.StorageReadOp{
		Name:   *f.name,
		Offset: f.offset,
		Count:  int64(len(p)),
	})
	if grpcErr != nil {
		return 0, grpcErr
	}
	copy(resp.Data, p)
	return len(resp.Data), storageError(&resp.Error)
}

func (f *file) Write(p []byte) (n int, err error) {
	resp, grpcErr := f.fs.client.Storage.Write(f.fs.ctx, &sourcegraph.StorageWriteOp{
		Name:   *f.name,
		Offset: f.offset,
		Data:   p,
	})
	if grpcErr != nil {
		return 0, grpcErr
	}
	return int(resp.Wrote), storageError(resp.Error)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if offset < 0 {
		panic("File.Seek: cannot seek to a negative offset")
	}
	switch whence {
	case 0:
		f.offset = offset
	case 1:
		f.offset += offset
	case 2:
		fi, err := f.fs.Lstat(f.name.Name)
		if err != nil {
			return 0, err
		}
		f.offset = fi.Size() - offset
	default:
		panic("File.Seek: invalid whence value")
	}
	return f.offset, nil
}

func (f *file) Close() error {
	ioErr, grpcErr := f.fs.client.Storage.Close(f.fs.ctx, f.name)
	if grpcErr != nil {
		return grpcErr
	}
	return storageError(ioErr)
}
