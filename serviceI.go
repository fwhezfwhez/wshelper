package wshelper

type ServiceI interface{
	SendOne()
	SendMany()
	SendGroup()
}
