package object

type blob struct {
	*object
	raw      []byte
	filename string
}

func (b *blob) Filename() string {
	return b.filename
}

type Blob interface {
	Object
	Filename() string
}

func NewBlob(filename string, data []byte) (Blob, error) {

	object, err := newObject(data, classBlob)
	if err != nil {
		return nil, err
	}

	return &blob{object, data, filename}, nil
}
