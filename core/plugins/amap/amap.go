package amap

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type AmapType struct {
	key string
}

const (
	const_URL_IP        = "https://restapi.amap.com/v3/ip?"                  // 查询IP的api
	const_URL_GEOCODE   = "https://restapi.amap.com/v3/geocode/geo?"         // 地理编码查询的api
	const_URL_REGEOCODE = "https://restapi.amap.com/v3/geocode/regeo?"       // 逆向地理编码查询的api
	const_URL_PLACE     = "https://restapi.amap.com/v3/place/text?"          // 搜索POI api
	const_URL_WEATHER   = "https://restapi.amap.com/v3/weather/weatherInfo?" // 查询天气 api
)

func NewAmap(key string) *AmapType {
	return &AmapType{key}
}

// 查询IP
type checkIpResType struct {
	Status    string `json:"status""`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Adcode    string `json:"adcode"`
	Rectangle string `json:"rectangle"`
}

func (_self *AmapType) CheckIp(IP string) checkIpResType {
	var result checkIpResType
	getUrl := const_URL_IP + "key=" + _self.key + "&ip=" + IP
	res, _ := http.Get(getUrl)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &result)

	return result
}

// 地址查询
type checkGeocodeResType struct {
	Status   string `json:"status"`
	Count    int64  `json:"count"`
	Info     string `json:"info"`
	Geocodes []struct {
		Formatted_address string `json:"formatted_address"`
		Province          string `json:"province"`
		City              string `json:"city"`
		Citycode          string `json:"citycode"`
		District          string `json:"district"`
		Township          string `json:"township"`
		Street            string `json:"street"`
		Number            string `json:"number"`
		Adcode            string `json:"adcode"`
		Location          string `json:"location"`
		Level             string `json:`
	} `json:"geocodes"`
}

func (_self *AmapType) CheckGeocode(address string) checkGeocodeResType {
	var result checkGeocodeResType
	url := const_URL_GEOCODE + "key=" + _self.key + "&address=" + address
	res, _ := http.Get(url)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &result)

	return result
}

// 经纬度查询
type checkRegeocodeResType struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Regeocode struct {
		Formatted_address string `json:"formattedAddress"`
		AddressComponent  struct {
			Country      string `json:"country"`
			Province     string `json:"province"`
			City         string `json:"city"`
			Citycode     string `json:"citycode"`
			District     string `json:"district"`
			Adcode       string `json:"adcode"`
			Township     string `json:"township"`
			Towncode     string `json:"towncode"`
			SeaArea      string `json:"seaArea"`
			Neighborhood struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"neighborhood"`
			Building struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"building"`
			StreetNumber struct {
				Street    string `json:"street"`
				Number    string `json:"number"`
				Location  string `json:"location"`
				Direction string `json:"direction"`
				Distance  string `json:"distance"`
			} `json:"streetNumber"`
			BusinessAreas []struct {
				Location string `json:"location"`
				Name     string `json:"name"`
				Id       string `json:"id"`
			} `json:"businessAreas"`
		} `json:"addressComponent"`
	} `json:"regeocode"`
}

func (_self *AmapType) CheckRegeocode(location string) checkRegeocodeResType {
	var result checkRegeocodeResType
	url := const_URL_REGEOCODE + "key=" + _self.key + "&location=" + location
	res, _ := http.Get(url)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &result)

	return result
}

// 搜索POI
// 地址查询
type checkPlaceKeyWordType struct {
	Status     string `json:"status"`
	Count      int64  `json:"count"`
	Info       string `json:"info"`
	Suggestion []struct {
		Keywords string `json:"keywords"`
		Cities   []struct {
			Name     string `json:"name"`
			Num      string `json:"num"`
			Citycode string `json:"citycode"`
			Adcode   string `json:"adcode"`
		} `json:"cities"`
	} `json:"suggestion"`
	Pois []struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		Type          string `json:"type"`
		Typecode      string `json:"typecode"`
		Biz_type      string `json:"biz_type"`
		Address       string `json:"address"`
		Location      string `json:"location"`
		Distance      string `json:"distance"`
		Tel           string `json:"tel"`
		Postcode      string `json:"postcode"`
		Website       string `json:"website"`
		Email         string `json:"email"`
		Pcode         string `json:"pcode"`
		Pname         string `json:"pname"`
		Citycode      string `json:"citycode"`
		Cityname      string `json:"cityname"`
		Adcode        string `json:"adcode"`
		Adname        string `json:"adname"`
		Entr_location string `json:"entr_location"`
		Exit_location string `json:"exit_location"`
		Navi_poiid    string `json:"navi_poiid"`
		Gridcode      string `json:"gridcode"`
		Alias         string `json:"alias"`
		Business_area string `json:"business_area"`
		Parking_type  string `json:"parking_type"`
		Tag           string `json:"tag"`
		Indoor_map    string `json:"indoor_map"`
	} `json:"pois"`
}

func (_self *AmapType) CheckPlaceForKeyWord(keywords string) checkPlaceKeyWordType {
	var result checkPlaceKeyWordType
	url := const_URL_PLACE + "key=" + _self.key + "&keywords=" + keywords
	res, _ := http.Get(url)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &result)

	return result
}

// 天气查询
type checkWeatherType struct {
	Status   string `json:"status""`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Count    string `json:"count"`

	Lives []struct {
		Province      string `json:"province"`
		City          string `json:"city"`
		Adcode        string `json:"adcode"`
		Weather       string `json:"weather"`
		Temperature   string `json:"temperature"`
		Winddirection string `json:"winddirection"`
		Windpower     string `json:"windpower"`
		Humidity      string `json:"humidity"`
		Reporttime    string `json:"reporttime"`
	} `json:"lives"`
}

func (_self *AmapType) CheckWeather(city string) checkWeatherType {
	var result checkWeatherType
	getUrl := const_URL_WEATHER + "key=" + _self.key + "&city=" + city
	res, _ := http.Get(getUrl)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &result)

	return result
}
