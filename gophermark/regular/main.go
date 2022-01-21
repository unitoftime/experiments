package main

import (
	"os"
	"strconv"
	"log"
	"time"
	"math/rand"
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/unitoftime/gl"
	"github.com/unitoftime/glfw"
	"github.com/unitoftime/gl/glutil"
)

//go:embed man.png
var f embed.FS

func loadImage(path string) (*image.NRGBA, error) {
	file, err := f.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

var width int = 1920
var height int = 1080

func run() error {
	// ------------
	// Setup Window
	// ------------
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	// glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True) // Compatibility - For Mac

	win, err := glfw.CreateWindow(width, height, "Benchmark", glfw.GetPrimaryMonitor(), nil)
	if err != nil {
		panic(err)
	}

	win.MakeContextCurrent()

	glfw.SwapInterval(0)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

	// ------------
	// load shader
	// ------------
	program, err := glutil.CreateProgram(vertexSource, fragmentSource)
	if err != nil {
		return err
	}

	mesh := NewMesh()

	// ------------
	// Load Texture
	// ------------
	manImage, err := loadImage("man.png")
	if err != nil {
		panic(err)
	}

	texture := gl.CreateTexture()
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	// Smoothing?
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D, 0, 160, 200, gl.RGBA, gl.UNSIGNED_BYTE, manImage.Pix)

	// runtime.SetFinalizer(&texture, func() {
	defer gl.DeleteTexture(texture)

	// ------------
	// Setup texture in shader
	// ------------
	gl.UseProgram(program)

	projMat := mgl32.Ortho2D(0, float32(width), 0, float32(height))
	projectionLoc := gl.GetUniformLocation(program, "projection")
	gl.UniformMatrix4fv(projectionLoc, projMat[:])

	// Note: Only necessary if multiple textures
	// textureLoc := gl.GetUniformLocation(program, "texture1")
	// gl.Uniform1i(textureLoc, 0)

	transformLoc := gl.GetUniformLocation(program, "transform")

	length, err := strconv.Atoi(os.Args[1])
	if err != nil { panic(err) }
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := float32(160.0)/4
	h := float32(200.0)/4

	start := time.Now()
	for i := 0; i < 1000; i++ {
		start = time.Now()

		// Physics
		for i := range man {
			man[i].position[0] += man[i].velocity[0]
			man[i].position[1] += man[i].velocity[1]

			if man[i].position[0] <= 0 || (man[i].position[0]+w) >= float32(width) {
				man[i].velocity[0] = -man[i].velocity[0]
			}
			if man[i].position[1] <= 0 || (man[i].position[1]+h) >= float32(height) {
				man[i].velocity[1] = -man[i].velocity[1]
			}
		}

		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.UseProgram(program)
		mesh.Bind()

		for i := range man {
			mat := man[i].Matrix()
			gl.UniformMatrix4fv(transformLoc, mat[:])

			mesh.Draw()
		}

		win.SwapBuffers()
		glfw.PollEvents()

		time.Sleep(16 * time.Millisecond)

		dt := time.Since(start)
		log.Println(dt.Seconds() * 1000)
	}

	// for i := range times {
	// 	log.Println(times[i].Count, 1000 * times[i].Time.Seconds())
	// }
	return err
}

type Frame struct {
	Count int
	Time time.Duration
}

type Man struct {
	position, velocity mgl32.Vec2
}
func NewMan() Man {
	vScale := 5.0
	return Man{
		// position: mgl32.Vec2{100, 100},
		// position: mgl32.Vec2{float32(float64(width/2) * rand.Float64()),
		// 	float32(float64(height/2) * rand.Float64())},
		position: mgl32.Vec2{1920/2, 1080/2},
		velocity: mgl32.Vec2{float32(2*vScale * (rand.Float64()-0.5)),
			float32(2*vScale * (rand.Float64()-0.5))},
	}
}

func (m *Man) Matrix() mgl32.Mat4 {
	return mgl32.Translate3D(m.position[0], m.position[1], 0)
}

type Mesh struct {
	vao, vbo, ebo gl.Buffer
}

func (m *Mesh) Bind() {
	gl.BindVertexArray(m.vao)
}

func (m *Mesh) Draw() {
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, 0)
}

func NewMesh() *Mesh {
	m := Mesh{}

	// If I flipped images:
	// vertices := []float32{
	// 	// positions          // colors           // texture coords
	// 	0.5		,  0.5, 0.0,   1.0, 0.0, 0.0,   1.0, 1.0, // top right
	// 	0.5		, -0.5, 0.0,   0.0, 1.0, 0.0,   1.0, 0.0, // bottom right
	// 	-0.5	, -0.5, 0.0,   0.0, 0.0, 1.0,   0.0, 0.0, // bottom left
	// 	-0.5	,  0.5, 0.0,   1.0, 1.0, 0.0,   0.0, 1.0,  // top left
	// };
	// Unflipped, not projected
	// vertices := []float32{
	// 	// positions          // colors           // texture coords
	// 	0.5		,  0.5, 0.0,   1.0, 0.0, 0.0,   1.0, 0.0, // top right
	// 	0.5		, -0.5, 0.0,   0.0, 1.0, 0.0,   1.0, 1.0, // bottom right
	// 	-0.5	, -0.5, 0.0,   0.0, 0.0, 1.0,   0.0, 1.0, // bottom left
	// 	-0.5	,  0.5, 0.0,   1.0, 1.0, 0.0,   0.0, 0.0,  // top left
	// };
	w := float32(160.0)/4
	h := float32(200.0)/4
	vertices := []float32{
		// positions          // colors           // texture coords
		w	,  h, 0.0,   1.0, 0.0, 0.0,   1.0, 0.0, // top right
		w	,  0, 0.0,   0.0, 1.0, 0.0,   1.0, 1.0, // bottom right
		0	,  0, 0.0,   0.0, 0.0, 1.0,   0.0, 1.0, // bottom left
		0	,  h, 0.0,   1.0, 1.0, 0.0,   0.0, 0.0,  // top left
	};

	indices := []uint32{
		0, 1, 3, // first triangle
		1, 2, 3,  // second triangle
	};

	m.vao = gl.GenVertexArrays()
	m.vbo = gl.GenBuffers()
	m.ebo = gl.GenBuffers()

	gl.BindVertexArray(m.vao)

	// vbo
	componentSize := 4 // float32
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, componentSize * len(vertices), vertices, gl.STATIC_DRAW)

	// ebo
	indexSize := 4 // uint32
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(indices), indices, gl.STATIC_DRAW)

	// Location 0 = Position
	gl.VertexAttribPointer(gl.Attrib{0}, 3, gl.FLOAT, false, 8 * componentSize, 0)
	gl.EnableVertexAttribArray(gl.Attrib{0})

	// Location 1 = color
	gl.VertexAttribPointer(gl.Attrib{1}, 3, gl.FLOAT, false, 8 * componentSize, 3 * componentSize)
	gl.EnableVertexAttribArray(gl.Attrib{1})

	// Location 2 = texture
	gl.VertexAttribPointer(gl.Attrib{2}, 2, gl.FLOAT, false, 8 * componentSize, 6 * componentSize)
	gl.EnableVertexAttribArray(gl.Attrib{2})

	return &m
}

const (
	vertexSource = `
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aColor;
layout (location = 2) in vec2 aTexCoord;

out vec3 ourColor;
out vec2 TexCoord;

uniform mat4 projection;
uniform mat4 transform;

void main()
{
	gl_Position = projection * transform * vec4(aPos, 1.0);
	ourColor = aColor;
	TexCoord = vec2(aTexCoord.x, aTexCoord.y);
}
`
	fragmentSource = `
#version 330 core
out vec4 FragColor;

in vec3 ourColor;
in vec2 TexCoord;

// texture samplers
uniform sampler2D texture1;

void main()
{
	// linearly interpolate between both textures (80% container, 20% awesomeface)
	//FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
  FragColor = texture(texture1, TexCoord);
}
`
)
