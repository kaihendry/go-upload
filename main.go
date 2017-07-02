package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/tajtiattila/metadata/exif"
)

func upload(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20) // Not quite sure what this should be

	r.ParseForm()

	fmt.Println("Lat:", r.Form["lat"])
	fmt.Println("Lng:", r.Form["lng"])
	lat, err := strconv.ParseFloat(r.Form["lat"][0], 64)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	lng, err := strconv.ParseFloat(r.Form["lng"][0], 64)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	file, _, err := r.FormFile("jpeg")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	filetype := http.DetectContentType(buff)
	fmt.Println(filetype)

	if filetype != "image/jpeg" {
		http.Error(w, "Upload not a JPEG", 400)
		return
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	x, err := exif.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	x.SetLatLong(lat, lng)

	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// f, err := ioutil.TempFile("", "upload")
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	w.Header().Add("Content-Type", "image/jpeg")

	if err := exif.Copy(w, file, x); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// if err := f.Close(); err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	// fmt.Printf("Upload written to %v\n", f.Name())
	// w.Write([]byte("OK"))
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/upload", upload)

	log.Println("Listening on :3001")
	err := http.ListenAndServe(":3001", mux)
	log.Fatal(err)

}

func index(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("foo").Parse(`<!DOCTYPE html>
<html>
<head>
<title>Simplest upload example</title>
<meta charset="utf-8" />
<meta name=viewport content="width=device-width, initial-scale=1">
</head>
<body>


<div id="test1">

<form action="/upload" enctype="multipart/form-data" method="post">
<input type="file" required name="jpeg" />
<br>
<label>Lat: <input type="number" step="any" name=lat required v-model.number.lazy="mapCenter.lat" /></label>
<label>Lng: <input type="number" step="any" name=lng required v-model.number.lazy="mapCenter.lng" /></label>
<input type="submit" value="Report" />
</form>

<gmap-map style="width: 500px; height: 500px" :center="mapCenter" :zoom="12"
  @center_changed="updateCenter"
  class="map-container">
</gmap-map>

</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/vue/2.2.0/vue.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.16.4/lodash.js"></script>
<script src="https://s.natalian.org/2017-06-28/vue-google-maps.js"></script>

  <script>
    Vue.use(VueGoogleMaps, {
      load: {
        key: 'AIzaSyD4VHBovJ2dnHSZpS-Y46hheA_JL6mtwZI',
      }
    });

    document.addEventListener('DOMContentLoaded', function() {
      new Vue({
        el: '#test1',
        data: {
          mapCenter: {
            lat: 1.38,
            lng: 103.8,
          }
		},
		created: function () {
			console.log('a is: ' + this.mapCenter)
			if(navigator.geolocation) {
				navigator.geolocation.getCurrentPosition((position) => {
					console.log(position.coords.latitude, position.coords.longitude)
					this.mapCenter.lat = position.coords.latitude
					this.mapCenter.lng = position.coords.longitude
				})
			}
		},
        methods: {
          updateCenter(newCenter) {
            this.mapCenter = {
              lat: newCenter.lat(),
              lng: newCenter.lng(),
            }
          }
        }
      });
    });
  </script>


</body>
</html>`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, t)

}
