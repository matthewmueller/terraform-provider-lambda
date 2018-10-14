package lambda

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/matthewmueller/terraform-provider-lambda/internal/archive"
	"github.com/matthewmueller/terraform-provider-lambda/internal/golang"
)

func resourceGo() *schema.Resource {
	return &schema.Resource{
		Create: resourceGoCreate,
		Read:   resourceGoRead,
		Update: resourceGoUpdate,
		Delete: resourceGoDelete,
		Schema: map[string]*schema.Schema{
			// TODO: support single files
			"source": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source of the lambda function",
			},
			"path": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Path to the output zip",
			},
			"base64sha256": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA1 checksum of the zip file.",
			},
			"md5": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MD5 checksum of the zip file.",
			},
			"size": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Size of the zip file.",
			},
		},
	}
}

func resourceGoCreate(d *schema.ResourceData, meta interface{}) error {
	source := d.Get("source").(string)

	// check to make sure the source exists
	if stat, err := os.Stat(source); err != nil {
		d.SetId("")
		return err
	} else if !stat.IsDir() {
		d.SetId("")
		return err
	}

	zip, err := compileGo(source)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(zip)
	if err != nil {
		return err
	}
	d.Set("size", len(buf))

	// compute the hashes
	d.Set("base64sha256", base64SHA256(buf))
	md5 := md5Hash(buf)
	d.Set("md5", md5)

	// emphemeral cache
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	path := filepath.Join(cache, "terraform-provider-lambda", md5+".zip")
	d.Set("path", path)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// write the file
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		return err
	}

	d.SetId(md5)
	return nil
}

func resourceGoRead(d *schema.ResourceData, meta interface{}) error {
	source := d.Get("source").(string)

	// check to make sure the source exists
	if stat, err := os.Stat(source); os.IsNotExist(err) {
		d.SetId("")
		return nil
	} else if err != nil {
		d.SetId("")
		return err
	} else if !stat.IsDir() {
		d.SetId("")
		return nil
	}

	zip, err := compileGo(source)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(zip)
	if err != nil {
		return err
	}
	d.Set("size", len(buf))

	// compute the hashes
	d.Set("base64sha256", base64SHA256(buf))
	md5 := md5Hash(buf)
	d.Set("md5", md5)

	oldPath := d.Get("path").(string)
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, md5+".zip")

	if oldPath != newPath {
		// remove the old path
		if err := rmrf(oldPath); err != nil {
			return err
		}

		// write to the new path
		if err := ioutil.WriteFile(newPath, buf, 0644); err != nil {
			return err
		}
	}

	d.Set("path", newPath)
	d.SetId(md5)
	return nil
}

func resourceGoUpdate(d *schema.ResourceData, meta interface{}) error {
	source := d.Get("source").(string)

	// check to make sure the source exists
	if stat, err := os.Stat(source); err != nil {
		d.SetId("")
		return err
	} else if !stat.IsDir() {
		d.SetId("")
		return err
	}

	zip, err := compileGo(source)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(zip)
	if err != nil {
		return err
	}
	d.Set("size", len(buf))

	// compute the hashes
	d.Set("base64sha256", base64SHA256(buf))
	md5 := md5Hash(buf)
	d.Set("md5", md5)

	oldPath := d.Get("path").(string)
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, md5+".zip")
	if oldPath != newPath {
		// remove the old path
		if err := rmrf(oldPath); err != nil {
			return err
		}

		// write to the new path
		if err := ioutil.WriteFile(newPath, buf, 0644); err != nil {
			return err
		}
	}

	d.Set("path", newPath)
	d.SetId(md5)
	return nil
}

func resourceGoDelete(d *schema.ResourceData, meta interface{}) error {
	path := d.Get("path").(string)
	return rmrf(path)
}

func base64SHA256(b []byte) string {
	h := sha256.New()
	h.Write(b)
	sha := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sha)
}

func md5Hash(b []byte) string {
	m := md5.New()
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

func compileGo(source string) (io.Reader, error) {
	env := make(map[string]string)
	env["GOOS"] = "linux"
	env["GOARCH"] = "amd64"
	env["GOPATH"] = os.Getenv("GOPATH")

	// compile the function to path
	mainpath := filepath.Join(source, "main")
	if err := golang.Compile(source, mainpath, env); err != nil {
		return nil, err
	}
	defer rmrf(mainpath)

	zip, _, err := archive.Zip(source)
	if err != nil {
		return nil, err
	}
	return zip, nil
}

// cleanup our function
func rmrf(filename string) error {
	if err := os.RemoveAll(filename); err != nil {
		return err
	}
	return nil
}
