package files

type FileHandler struct{}

// Max Size of TXT
const MAX_CHUNK_SIZE = 1

type FileHandlerUploadResponse struct {
	FilePath string `json:"file_path"` // txt record of index
}

func (f *FileHandler) Upload(filePath string) error {
	// Canculate the no of chunks need to create
	// Create a index TXT record for that chuck file
	// NAME:<FILE_TYPE>:<START_CHUNK|0>:<END_CHUNK|100>

	// Start reading the file into MAX_CHUNK_SIZE byte and add its txt record (io.pipe)
	panic("implement me")
	return nil
}

func (f *FileHandler) Download(record string) error {
	// Check correct index file record
	// If correct, read the TXT records and pipe data into that file
	panic("implement me")
	return nil
}

func (f *FileHandler) Stream(record string) error {
	// I do not know it
	panic("implement me")
	return nil
}

func (f *FileHandler) Delete(record string) error {
	// Check correct index file record
	// Delete those TXT record
	panic("implement me")
	return nil
}
