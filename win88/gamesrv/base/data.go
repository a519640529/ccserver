package base

type Data struct {
	datas map[string]interface{}
}

func NewData() *Data {
	return &Data{datas: make(map[string]interface{})}
}

func (d *Data) SetData(key string, val interface{}) {
	d.datas[key] = val
}

func (d *Data) GetData(key string) interface{} {
	if v, exist := d.datas[key]; exist {
		return v
	}
	return nil
}

type InterventionData struct {
	Webuser    string
	Flag       int32
	NumOfGames int32
}

type InterventionResults struct {
	Key     string
	Webuser string
	Results string
}
