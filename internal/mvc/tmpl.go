package mvc

import (
	"encoding/json"
	"fmt"
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	"go-auth/internal/tool/toolstring"
	xweb "go-auth/internal/web"
	"html/template"
	"io"
	"io/fs"
	"time"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer interface {
	Render(w io.Writer, name string, data any, c echo.Context) error
}

type templateRenderer struct {
	templates *template.Template
}

type ModelAppConfig struct {
	AppTitle              string
	CopyrightTitle        string
	AssetsGlobalVersion   string
	AssetsPublicURLOrigin string
	AssetsPathIndexCSS    string
	AssetsPathIndexJs     string
}

type ModelConst struct {
	SecretCodeLength     int // 6
	PasswordMinLength    int // 8
	DefaultTextSize      int // 100
	PhoneNumberMinLength int // 8
	PhoneNumberMaxLength int // 14
}
type ModelPrm struct {
	Title      string
	AppTitle   string
	Csrf       string
	IsFragment bool

	LangCode    string
	AppConfig   ModelAppConfig
	AppConst    ModelConst
	RawJSONData any
}

type ModelAPI struct {
	IsAuthenticated bool
	Icon            func(string) template.HTML
	Lang            func(text string, args ...any) string
	URL             func(path string, args ...string) string // path?args[0]=args[1]&args[2]=args[3]#args[4]
}

type ModelWrap struct {
	Model any
	Prm   ModelPrm
	API   ModelAPI
}

func NewModelWrap(c echo.Context, model any, isFragment bool, title string, appConfig *config.AppConfig, lang UserLang) (*ModelWrap, error) {

	_csrf, _ := c.Get("_csrf").(string)

	res := &ModelWrap{
		Model: model,
		API: ModelAPI{
			Icon:            AppIcons,
			Lang:            lang.Lang,
			URL:             toolstring.LocalURL,
			IsAuthenticated: xweb.IsSignedIn(c),
		},
		Prm: ModelPrm{

			Title: title,

			Csrf: _csrf,

			IsFragment: isFragment,

			LangCode: lang.LangCode(),

			AppConfig: ModelAppConfig{
				AppTitle:              appConfig.Title,
				CopyrightTitle:        fmt.Sprintf("© %v %v", time.Now().UTC().Year(), appConfig.Title),
				AssetsGlobalVersion:   appConfig.Assets.AssetsGlobalVersion,
				AssetsPublicURLOrigin: appConfig.Assets.AssetsPublicURLOrigin,
				AssetsPathIndexCSS:    appConfig.Assets.AssetsPathIndexCSS,
				AssetsPathIndexJs:     appConfig.Assets.AssetsPathIndexJs,
			},
			AppConst: ModelConst{
				SecretCodeLength:     consts.SecretCodeLength,
				PasswordMinLength:    consts.PasswordMinLength,
				DefaultTextSize:      consts.DefaultTextSize, // 100
				PhoneNumberMinLength: consts.PhoneNumberMinLength,
				PhoneNumberMaxLength: consts.PhoneNumberMaxLength,
			},
		},
	}

	data, err := json.Marshal(map[string]any{
		"test":                     `"<>`,
		"PrmLangCode":              res.Prm.LangCode,
		"PrmAssetsPublicURLOrigin": res.Prm.AppConfig.AssetsPublicURLOrigin,
		"PrmAssetsGlobalVersion":   res.Prm.AppConfig.AssetsGlobalVersion,
	})

	if err != nil {
		return nil, err
	}

	res.Prm.RawJSONData = data
	return res, nil
}

func NewTemplateRenderer(viewsFs fs.FS, patterns ...string) TemplateRenderer {

	res := templateRenderer{}
	res.templates = template.Must(template.ParseFS(viewsFs, patterns...))

	return &res
}

// Render renders a template document
func (x *templateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	//
	isFragment := false /*use global(wrap) layout*/

	if model, ok := data.(*ModelWrap); ok {
		isFragment = model.Prm.IsFragment
	}

	if !isFragment {
		err := x.templates.ExecuteTemplate(w, "layout_header", data)

		if err != nil {
			return err
		}
	}

	{

		err := x.templates.ExecuteTemplate(w, name, data)
		if err != nil {
			return err
		}
	}

	if !isFragment {
		err := x.templates.ExecuteTemplate(w, "layout_footer", data)
		if err != nil {
			return err
		}
	}

	return nil
}
