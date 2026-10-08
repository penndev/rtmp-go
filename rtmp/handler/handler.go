package handler

type Handler interface {
	OnName(app, stream string) string

	OnPublish(app, stream string) bool
	OnPublishStop(app, stream string)

	OnPlay(app, stream string) bool
	OnPlayStop(app, stream string)
}
