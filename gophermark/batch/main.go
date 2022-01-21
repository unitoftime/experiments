package main

import (
	"os"
	"strconv"
	"fmt"
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
		panic(err)
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

	// win, err := glfw.CreateWindow(width, height, "Benchmark", glfw.GetPrimaryMonitor(), nil)
	win, err := glfw.CreateWindow(width, height, os.Args[1], nil, nil)
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

	length, err := strconv.Atoi(os.Args[1])
	if err != nil { panic(err) }

	numVerts := 4 * length
	numTris := 2 * length
	batch := NewBatch(numVerts, numTris)

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

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D, 0, 160, 200, gl.RGBA, gl.UNSIGNED_BYTE, manImage.Pix)

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
	identMat := []float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	gl.UniformMatrix4fv(transformLoc, identMat)

	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := float32(160.0)/4
	h := float32(200.0)/4

	start := time.Now()
	for i := 0; i < 10000; i++ {
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
		batch.Bind()

		batch.Clear()
		for i := range man {
			batch.Add(float32(man[i].position[0]), float32(man[i].position[1]),
				man[i].R, man[i].G, man[i].B)
		}

		batch.Draw()


		win.SwapBuffers()
		glfw.PollEvents()

		// time.Sleep(16 * time.Millisecond)

		dt := time.Since(start)
		fmt.Println(dt.Seconds() * 1000)
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
	R, G, B float32
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
		R: rand.Float32(),
		G: rand.Float32(),
		B: rand.Float32(),
	}
}

func (m *Man) Matrix() mgl32.Mat4 {
	return mgl32.Translate3D(m.position[0], m.position[1], 0)
}


type Batch struct {
	vao, vbo, ebo gl.Buffer
	// indexCount int
	// maxTri int
	// maxVert int
	vertices []float32
	indices []uint32
}

func (b *Batch) Bind() {
	gl.BindVertexArray(b.vao)
}

func (b *Batch) Clear() {
	b.vertices = b.vertices[:0]
	b.indices = b.indices[:0]
}
func (b *Batch) Draw() {
	gl.BindVertexArray(b.vao)

	// componentSize := 4 // float32
	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	// gl.BufferData(gl.ARRAY_BUFFER, componentSize * len(b.vertices), b.vertices, gl.DYNAMIC_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, b.vertices)

	// indexSize := 4 // uint32
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
	// gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(b.indices), b.indices, gl.DYNAMIC_DRAW)
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, b.indices)

	// fmt.Println(len(b.indices))
	gl.DrawElements(gl.TRIANGLES, len(b.indices), gl.UNSIGNED_INT, 0)
}

func (b *Batch) Add(x, y float32, R, G, B float32) {
	currentElement := uint32(len(b.vertices)/8) // Note: 8 b/c stride is 8

	w := float32(160.0)/4
	h := float32(200.0)/4
	b.vertices = append(b.vertices, []float32{
		// positions       // colors           // texture coords
		x+w	,  y+h, 0.0,   R, G, B,   1.0, 0.0, // top right
		x+w	,  y+0, 0.0,   R, G, B,   1.0, 1.0, // bottom right
		x+0	,  y+0, 0.0,   R, G, B,   0.0, 1.0, // bottom left
		x+0	,  y+h, 0.0,   R, G, B,   0.0, 0.0,  // top left
	}...)

	b.indices = append(b.indices, []uint32{
		currentElement + 0, currentElement + 1, currentElement + 3, // first triangle
		currentElement + 1, currentElement + 2, currentElement + 3,  // second triangle
	}...)

	// fmt.Println(currentElement, x, y, w, h, len(b.vertices), len(b.indices))
}

func NewBatch(numVerts, numTris int) *Batch {
	b := Batch{
		// maxVert: maxVerts,
		// maxTri: maxTriangles,
		vertices: make([]float32, 8 * 4 * numVerts), // 8 * floats (sizeof float) * num triangles
		indices: make([]uint32, 3 * numTris), // 3 indices per triangle
	}

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

	// w := float32(160.0)/4
	// h := float32(200.0)/4
	// vertices := []float32{
	// 	// positions          // colors           // texture coords
	// 	w	,  h, 0.0,   1.0, 0.0, 0.0,   1.0, 0.0, // top right
	// 	w	,  0, 0.0,   0.0, 1.0, 0.0,   1.0, 1.0, // bottom right
	// 	0	,  0, 0.0,   0.0, 0.0, 1.0,   0.0, 1.0, // bottom left
	// 	0	,  h, 0.0,   1.0, 1.0, 0.0,   0.0, 0.0,  // top left
	// };

	// indices := []uint32{
	// 	0, 1, 3, // first triangle
	// 	1, 2, 3,  // second triangle
	// };

	b.vao = gl.GenVertexArrays()
	b.vbo = gl.GenBuffers()
	b.ebo = gl.GenBuffers()

	gl.BindVertexArray(b.vao)

	// vbo
	componentSize := 4 // float32
	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, componentSize * len(b.vertices), b.vertices, gl.DYNAMIC_DRAW)

	// ebo
	indexSize := 4 // uint32
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(b.indices), b.indices, gl.DYNAMIC_DRAW)

	// Location 0 = Position
	gl.VertexAttribPointer(gl.Attrib{0}, 3, gl.FLOAT, false, 8 * componentSize, 0)
	gl.EnableVertexAttribArray(gl.Attrib{0})

	// Location 1 = color
	gl.VertexAttribPointer(gl.Attrib{1}, 3, gl.FLOAT, false, 8 * componentSize, 3 * componentSize)
	gl.EnableVertexAttribArray(gl.Attrib{1})

	// Location 2 = texture
	gl.VertexAttribPointer(gl.Attrib{2}, 2, gl.FLOAT, false, 8 * componentSize, 6 * componentSize)
	gl.EnableVertexAttribArray(gl.Attrib{2})

	return &b
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
  FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
}
`
)
