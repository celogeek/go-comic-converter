package epubprogress

import (
	"encoding/json"
)

type epubProgressJson struct {
	o       Options
	e       *json.Encoder
	current int
}

func (p *epubProgressJson) Add(num int) error {
	p.current += num
	p.e.Encode(map[string]any{
		"type": "progress",
		"data": map[string]any{
			"progress": map[string]any{
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
	return nil
}

func (p *epubProgressJson) Close() error {
	return nil
}
