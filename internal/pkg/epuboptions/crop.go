package epuboptions

type Crop struct {
	Enabled                 bool
	Left, Up, Right, Bottom int
	Limit                   int
	SkipIfLimitReached      bool
}
