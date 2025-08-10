package ads

type AdResponse struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	TargetURL string `json:"target_url"`
	Status    string `json:"status"`
}
