package minicommerce

// Downloadable is the location of a downloadable digital product uploaded somewhere to google cloud storage
type Downloadable struct {
	ID       string `firestore:"id,omitempty"`
	Location string `firestore:"location,omitempty"`
}
