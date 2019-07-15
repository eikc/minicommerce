package minicommerce

// Downloadable is the location of a downloadable digital product uploaded somewhere to google cloud storage
type Downloadable struct {
	ID       string `firestore:"-"`
	Name     string `firestore:"name,omitempty"`
	Location string `firestore:"location,omitempty"`
	Deleted  int64  `firestore:"deleted,omitempty"`
}
