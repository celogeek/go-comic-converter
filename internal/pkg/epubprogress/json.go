package epubprogress

import (
	"encoding/json"
)

type jsonprogress struct {
	o       Options
	e       *json.Encoder
	current int
}

func (p *jsonprogress) Add(num int) error {
	p.current += num
	return p.e.Encode(map[string]any{
		"type": "epubprogress",
		"data": map[string]any{
			"epubprogress": map[string]any{
				"current": p.current,
				"total":   p.o.Max,
			},
			"steps": map[string]any{
				"current": p.o.CurrentJob,
				"total":   p.o.TotalJob,
			},
			"description": p.o.Description,
		},
	})
}

func (p *jsonprogress) Close() error {
	return nil
}
