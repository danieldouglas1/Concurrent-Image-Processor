/*************************************************
 * Authors: Daniel Douglas
 * Date: May 25, 2020
 * Purpose: - Creates a web-server where users can upload an image
 *          - Does three crops to an image: Centre crop,
 *          - Resizes the image to 100x100 pixels
 *          - Converts the image to greyscale
 *          - Converts the image to ASCII Art
 *          - Logs the time for each function.
 *************************************************/

package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    "os"
    "path/filepath"
    "runtime"
    "log"
    "strings"
    "github.com/disintegration/imaging"
    "github.com/nfnt/resize"
    "flag"
    "bytes"
    "image"
    "reflect"
    "image/color"
    _ "image/png"
    _ "image/jpeg"
    "image/jpeg"
)


//Uploads file to the server
func uploadFile(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    time.Sleep(time.Second * 2)

    // Parse our multipart form, 10 << 20 specifies a maximum
    // upload of 10 MB files.
    r.ParseMultipartForm(10 << 20)

    // FormFile returns the first file for the given key `myFile`
    // it also returns the FileHeader so we can get the Filename,
    // the Header and the size of the file
    file, handler, err := r.FormFile("myFile")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()

    // Create a temporary file within our temp-images directory that follows
    // a particular naming pattern
    tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
    if err != nil {
    }
    defer tempFile.Close()

    // read all of the contents of our uploaded file into a
    // byte array
    fileBytes, err := ioutil.ReadAll(file)
    if err != nil {
        fmt.Println(err)
    }
    // write this byte array to our temporary file
    tempFile.Write(fileBytes)

    // return that we have successfully uploaded our file!
    fmt.Fprintf(w, "Successfully Uploaded File\n")
    fmt.Fprintf(w, "Uploaded File: %+v\n", handler.Filename)
    fmt.Fprintf(w, "File Size: %+v bytes\n", handler.Size)
    fmt.Fprintf(w, "\n", )
    fmt.Fprintf(w, "Running Image Function Stats: \n")

    midCrop(w,handler.Filename)
    greyScale(w, handler.Filename)
    resizePic(w, handler.Filename)
    toASCII(w, handler.Filename)

    elapsed := time.Since(start)
    fmt.Fprintf(w, "Program took %s\n", elapsed)
}

//Retreives the path of the URL
func setupRoutes() {
    http.HandleFunc("/upload", uploadFile)
    http.ListenAndServe(":8080", nil)
}

func main() {
    fmt.Println("Hello World")
    setupRoutes()
}

//Crops picture.
//1st Crop is a center crop of 600x 600 pixels
//2nd crop is a rectangular crop of
func midCrop(w http.ResponseWriter, name string){

  fmt.Fprint(w, "Cropping Image... \n")
  start := time.Now()
  time.Sleep(time.Second * 2)

  // load original image
  img, err := imaging.Open(name)

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  // crop from center
  centercropimg := imaging.CropCenter(img, 400, 400)

  // save cropped image
  err = imaging.Save(centercropimg, "./editImages/centercrop.jpg")

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  // crop out a rectangular region
  rectcropimg := imaging.Crop(img, image.Rect(0, 0, 600, 600))

  // save cropped image
  err = imaging.Save(rectcropimg, "./editImages/rectcrop.jpg")
  //error check
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  elapsed := time.Since(start)
  fmt.Fprintf(w, "Cropping took %s \n", elapsed)

}

//Resizes pictures to 100x100 pixels
func resizePic(w http.ResponseWriter, name string){
  fmt.Fprint(w, "Resizing Image...\n")
  start := time.Now()
  time.Sleep(time.Second * 2)

  runtime.GOMAXPROCS(runtime.NumCPU())

  img, err := imaging.Open(name)

  if err != nil {
    fmt.Println(err)
    os.Exit(1)

    elapsed := time.Since(start)
    fmt.Printf("Resize took %s", elapsed)
  }

  // resize image
  smallimg := imaging.Resize(img, 100, 100, imaging.NearestNeighbor)

  // save resized image in editedImages file
  err = imaging.Save(smallimg, "./editImages/smallpicture.jpg")

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

 elapsed := time.Since(start)
 fmt.Fprintf(w, "Resize took %s \n", elapsed)

}

//err checking for greyscale
func check(err error) {
    if err != nil {
        panic(err)
    }
}
//Converts image to black and white
func greyScale(w http.ResponseWriter, fileName string){

  fmt.Fprintf(w, "Converting Image to Grayscale...\n")
  start := time.Now()
  time.Sleep(time.Second * 2)

  imgPath := fileName
  f, err := os.Open(imgPath)

  check(err)
  defer f.Close()

  img, format, err := image.Decode(f)
  check(err)

  // only jpeg images can be converted
  if format != "jpeg" { log.Fatalln("Only jpeg/jpg images are supported") }
  size := img.Bounds().Size()
  rect := image.Rect(0, 0, size.X, size.Y)
  wImg := image.NewRGBA(rect)

  // loop though all the x
  for x := 0; x < size.X; x++ {
    // and now loop thorough all of this x's y
    for y := 0; y < size.Y; y++ {
      pixel := img.At(x, y)
      originalColor := color.RGBAModel.Convert(pixel).(color.RGBA)

      // Offset colors a little
      r := float64(originalColor.R) * 0.92126
      g := float64(originalColor.G) * 0.97152
      b := float64(originalColor.B) * 0.90722
      // average
      grey := uint8((r + g + b) / 3)
      c := color.RGBA{
      R: grey, G: grey, B: grey, A: originalColor.A,
      }
    wImg.Set(x, y, c)
    }
  }

  ext := filepath.Ext(imgPath)
  name := strings.TrimSuffix(filepath.Base(imgPath), ext)
  dir := "editImages"
  newImagePath := fmt.Sprintf("./%s/%s_gray%s", dir , name, ext)

  fg, err := os.Create(newImagePath)
  defer fg.Close()
  check(err)

  err = jpeg.Encode(fg, wImg, nil)
  check(err)

  elapsed := time.Since(start)
  fmt.Fprintf(w, "Greyscale took %s \n", elapsed)
}

// Characters included in ASCII image
var ASCIISTR = "MND8OZ$7I?+=~:,.."

// ASCII art helper function
func toASCII(w http.ResponseWriter, fileName string){
 fmt.Fprintf(w, "Converting to ASCII... \n")

 start := time.Now()
 time.Sleep(time.Second * 2)

  p := Convert2Ascii(ScaleImage(Init(fileName)))
  fmt.Print(string(p))

  elapsed := time.Since(start)
  fmt.Fprintf(w, "ASCII Conversion took %s \n", elapsed)

}

// ASCII art helper function
func Init(fileName string) (image.Image, int) {
  width := flag.Int("w", 80, "Use -w <width>")
  fpath := flag.String("p", fileName, "Use -p <filesource>")
  flag.Parse()

  f, err := os.Open(*fpath)
  if err != nil {
      log.Fatal(err)
  }
    img, _, err := image.Decode(f)
    if err != nil {
        log.Fatal(err)
    }

    f.Close()
    return img, *width
}

// ASCII art helper function
func ScaleImage(img image.Image, w int) (image.Image, int, int) {
    sz := img.Bounds()
    h := (sz.Max.Y * w * 10) / (sz.Max.X * 16)
    img = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
    return img, w, h
}

//Converts image to ASCII art
func Convert2Ascii(img image.Image, w, h int) []byte {
    table := []byte(ASCIISTR)
    buf := new(bytes.Buffer)

    for i := 0; i < h; i++ {
        for j := 0; j < w; j++ {
            g := color.GrayModel.Convert(img.At(j, i))
            y := reflect.ValueOf(g).FieldByName("Y").Uint()
            pos := int(y * 16 / 255)
            _ = buf.WriteByte(table[pos])
        }
        _ = buf.WriteByte('\n')
    }
    return buf.Bytes()
}
