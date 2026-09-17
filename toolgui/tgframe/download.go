package tgframe

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// Download is a file the page has offered the app user, and the token the
// client fetches it by. The bytes stay where the build keeps files -- on disk
// on a server, in the origin private file system in a tab -- so the pack
// carries the token and nothing else, rather than the file as base64.
type Download struct {
	token string
	name  string
	mime  string
	file  *File

	// sum is what the file holds. A page redraws its components on every run,
	// and one that offers the same bytes again must not rewrite the file or
	// rotate the token the last pack carried.
	sum [sha256.Size]byte
}

// Token is what the client fetches the file by. It is unguessable and it is
// this state's alone: no other state's store will find it.
func (d *Download) Token() string {
	return d.token
}

// Name is the filename the file is offered under.
func (d *Download) Name() string {
	return d.name
}

// MIME is the content type the file is served as.
func (d *Download) MIME() string {
	return d.mime
}

// Size is how many bytes the file holds.
func (d *Download) Size() int64 {
	return d.file.Size()
}

// Open return a reader over the content, for a transport that serves the
// file itself. The caller closes it.
func (d *Download) Open() (FileReader, error) {
	fp, err := d.file.Open()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return fp, nil
}

// downloadStore is one state's downloads, kept by the component that offered
// each one and by the token it is fetched with. A cloned state shares one, so
// what a run offers is what the transport serving that state can find.
type downloadStore struct {
	lock    sync.Mutex
	byOwner map[string]*Download
	byToken map[string]*Download
}

func newDownloadStore() *downloadStore {
	return &downloadStore{
		byOwner: make(map[string]*Download),
		byToken: make(map[string]*Download),
	}
}

// set offers bs under owner, which is the id of the component offering it, and
// returns the download to put in the pack.
//
// A rerun that offers the same file again gets the same download back, bytes
// and token both: the component's pack then doesn't change, and the token the
// client already holds goes on working. Different bytes are a different file
// and a different token, and the one before them stops being fetchable.
func (s *downloadStore) set(files *fileStore, owner, name, mime string,
	bs []byte) (*Download, error) {
	sum := sha256.Sum256(bs)

	s.lock.Lock()
	defer s.lock.Unlock()

	if cur := s.byOwner[owner]; cur != nil && cur.sum == sum &&
		cur.name == name && cur.mime == mime {
		return cur, nil
	}

	file, err := files.newFile(name)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if err := file.write(bytes.NewReader(bs), false); err != nil {
		file.body.remove()
		return nil, tgutil.Errorf("%w", err)
	}

	// A new file rather than the old one rewritten: a fetch already reading
	// the file this replaces reads all of what it opened, the way a reader of
	// an unlinked file does, instead of the bytes of a run it knows nothing
	// about.
	s.drop(owner)

	download := &Download{
		token: rand.Text(),
		name:  name,
		mime:  mime,
		file:  file,
		sum:   sum,
	}

	s.byOwner[owner] = download
	s.byToken[download.token] = download

	return download, nil
}

// remove drops what owner offered, its token and its bytes with it. It is how
// a download is released when the component offering it leaves the page for
// good: the file can be as large as the page liked, and waiting for the state
// to end would hold it against the disk or the origin's quota until then.
func (s *downloadStore) remove(owner string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.drop(owner)
}

// drop forgets what owner offered and takes its bytes with it. It must be
// called with the lock held.
func (s *downloadStore) drop(owner string) {
	old := s.byOwner[owner]
	if old == nil {
		return
	}

	delete(s.byOwner, owner)
	delete(s.byToken, old.token)

	old.file.body.remove()
}

func (s *downloadStore) get(token string) *Download {
	if token == "" {
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.byToken[token]
}
