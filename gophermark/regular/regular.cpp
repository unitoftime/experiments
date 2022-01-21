// This is adapted from LearnOpenGl.com
// This won't compile, but you can poke around to see what I did.


#include <glad/glad.h>
#include <GLFW/glfw3.h>
#include <stb_image.h>

#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>

#include <learnopengl/filesystem.h>
#include <learnopengl/shader_s.h>

#include <iostream>

#include <chrono>
using namespace std::chrono;

void framebuffer_size_callback(GLFWwindow* window, int width, int height);
void processInput(GLFWwindow *window);

// settings
const unsigned int SCR_WIDTH = 1920;
const unsigned int SCR_HEIGHT = 1080;

struct Man {
  glm::vec2 position;
  glm::vec2 velocity;
};


int main(int argc, char* argv[])
{
    // glfw: initialize and configure
    // ------------------------------
    glfwInit();
    glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 3);
    glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 3);
    glfwWindowHint(GLFW_OPENGL_PROFILE, GLFW_OPENGL_CORE_PROFILE);

// #ifdef __APPLE__
//     glfwWindowHint(GLFW_OPENGL_FORWARD_COMPAT, GL_TRUE);
// #endif

    // glfw window creation
    // --------------------
    GLFWwindow* window = glfwCreateWindow(SCR_WIDTH, SCR_HEIGHT, "Benchmark", NULL, NULL);
    if (window == NULL)
    {
        std::cout << "Failed to create GLFW window" << std::endl;
        glfwTerminate();
        return -1;
    }
    glfwMakeContextCurrent(window);
    // glfwSetFramebufferSizeCallback(window, framebuffer_size_callback);

    // glad: load all OpenGL function pointers
    // ---------------------------------------
    if (!gladLoadGLLoader((GLADloadproc)glfwGetProcAddress))
    {
        std::cout << "Failed to initialize GLAD" << std::endl;
        return -1;
    }

    glEnable(GL_BLEND);
    glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);

    glfwSwapInterval(0);

    // build and compile our shader zprogram
    // ------------------------------------
    Shader ourShader("4.1.texture.vs", "4.1.texture.fs"); 

    // set up vertex data (and buffer(s)) and configure vertex attributes
    // ------------------------------------------------------------------
    float w = 160/4;
    float h = 200/4;
    float vertices[] = {
        // positions          // colors           // texture coords
         w,  h, 0.0f,   1.0f, 0.0f, 0.0f,   1.0f, 0.0f, // top right
         w, 0.0f, 0.0f,   0.0f, 1.0f, 0.0f,   1.0f, 1.0f, // bottom right
        0.0f, 0.0f, 0.0f,   0.0f, 0.0f, 1.0f,   0.0f, 1.0f, // bottom left
        0.0f,  h, 0.0f,   1.0f, 1.0f, 0.0f,   0.0f, 0.0f  // top left 
    };
    unsigned int indices[] = {  
        0, 1, 3, // first triangle
        1, 2, 3  // second triangle
    };
    unsigned int VBO, VAO, EBO;
    glGenVertexArrays(1, &VAO);
    glGenBuffers(1, &VBO);
    glGenBuffers(1, &EBO);

    glBindVertexArray(VAO);

    glBindBuffer(GL_ARRAY_BUFFER, VBO);
    glBufferData(GL_ARRAY_BUFFER, sizeof(vertices), vertices, GL_STATIC_DRAW);

    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, EBO);
    glBufferData(GL_ELEMENT_ARRAY_BUFFER, sizeof(indices), indices, GL_STATIC_DRAW);

    // position attribute
    glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, 8 * sizeof(float), (void*)0);
    glEnableVertexAttribArray(0);
    // color attribute
    glVertexAttribPointer(1, 3, GL_FLOAT, GL_FALSE, 8 * sizeof(float), (void*)(3 * sizeof(float)));
    glEnableVertexAttribArray(1);
    // texture coord attribute
    glVertexAttribPointer(2, 2, GL_FLOAT, GL_FALSE, 8 * sizeof(float), (void*)(6 * sizeof(float)));
    glEnableVertexAttribArray(2);


    // load and create a texture 
    // -------------------------
    unsigned int texture;
    glGenTextures(1, &texture);
    glBindTexture(GL_TEXTURE_2D, texture); // all upcoming GL_TEXTURE_2D operations now have effect on this texture object
    // set the texture wrapping parameters
    // glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT);	// set texture wrapping to GL_REPEAT (default wrapping method)
    // glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);	// set texture wrapping to GL_REPEAT (default wrapping method)
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);
    // set texture filtering parameters
    // glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR_MIPMAP_LINEAR);
    // glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST);
    // load image, create texture and generate mipmaps
    int width, height, nrChannels;
    // The FileSystem::getPath(...) is part of the GitHub repository so we can find files on any IDE/platform; replace it with your own image path.
    unsigned char *data = stbi_load(FileSystem::getPath("resources/textures/man.png").c_str(), &width, &height, &nrChannels, 0);
    if (data)
    {
        glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, width, height, 0, GL_RGBA, GL_UNSIGNED_BYTE, data);
        // glGenerateMipmap(GL_TEXTURE_2D);
    }
    else
    {
        std::cout << "Failed to load texture" << std::endl;
    }
    stbi_image_free(data);

    ourShader.use();
    // glm::mat4 projMat = glm::ortho(0, glm::float32(SCR_WIDTH), 0, glm::float32(SCR_HEIGHT));
    glm::mat4 projMat = glm::mat4(1.0f);
    projMat = glm::ortho(float(0.0), float(SCR_WIDTH), float(0.0), float(SCR_HEIGHT));

    int projLoc = glGetUniformLocation(ourShader.ID, "projection");
    glUniformMatrix4fv(projLoc, 1, GL_FALSE, glm::value_ptr(projMat));

    int transformLoc = glGetUniformLocation(ourShader.ID, "transform");
    glm::mat4 transformMat = glm::mat4(1.0f);
    glUniformMatrix4fv(transformLoc, 1, GL_FALSE, glm::value_ptr(transformMat));

    float vScale = 5.0f;
    srand(1);
    char *p;
    int length = strtol(argv[1], &p, 10);
    Man man [length];
    for (int i=0; i<length; i++) {
      man[i].position.x = rand() % SCR_WIDTH;
      man[i].position.y = rand() % SCR_HEIGHT;
      man[i].velocity.x = vScale * float(rand() % 100)/100.0f;
      man[i].velocity.y = vScale * float(rand() % 100)/100.0f;

    }
    // man.position.x = 100;
    // man.position.y = 100;

    // render loop
    // -----------
    // while (!glfwWindowShouldClose(window))
    for(int i=0; i<10000; i++) {
      auto start = high_resolution_clock::now();
        // input
        // -----
      // processInput();

        // Physics
        // -----
        for (int i=0; i<length; i++) {
          man[i].position.x += man[i].velocity.x;
          man[i].position.y += man[i].velocity.y;

          if (man[i].position.x <= 0 || man[i].position.x >= float(SCR_WIDTH)) {
            man[i].velocity.x = -man[i].velocity.x;
          }
          if (man[i].position.y <= 0 || man[i].position.y >= float(SCR_HEIGHT)) {
            man[i].velocity.y = -man[i].velocity.y;
          }
        }

        // render
        // ------
        glClearColor(0.0f, 0.0f, 0.0f, 1.0f);
        // glClearColor(0.2f, 0.3f, 0.3f, 1.0f);
        glClear(GL_COLOR_BUFFER_BIT);

        // bind Texture
        glBindTexture(GL_TEXTURE_2D, texture);

        // render container
        ourShader.use();
        glBindVertexArray(VAO);

        for (int i=0; i<length; i++) {
          glm::mat4 transformMat = glm::mat4(1.0f);
          transformMat = glm::translate(transformMat, glm::vec3(man[i].position.x, man[i].position.y, 0.0f));
          glUniformMatrix4fv(transformLoc, 1, GL_FALSE, glm::value_ptr(transformMat));
          glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, 0);
        }

        // glfw: swap buffers and poll IO events (keys pressed/released, mouse moved etc.)
        // -------------------------------------------------------------------------------
        glfwSwapBuffers(window);
        glfwPollEvents();
        auto stop = high_resolution_clock::now();
        auto duration = duration_cast<nanoseconds>(stop - start);

        std::cout << float(duration.count())/1000000.0f << std::endl;
    }

    // optional: de-allocate all resources once they've outlived their purpose:
    // ------------------------------------------------------------------------
    glDeleteVertexArrays(1, &VAO);
    glDeleteBuffers(1, &VBO);
    glDeleteBuffers(1, &EBO);

    // glfw: terminate, clearing all previously allocated GLFW resources.
    // ------------------------------------------------------------------
    glfwTerminate();
    return 0;
}

// process all input: query GLFW whether relevant keys are pressed/released this frame and react accordingly
// ---------------------------------------------------------------------------------------------------------
void processInput(GLFWwindow *window)
{
    if (glfwGetKey(window, GLFW_KEY_BACKSPACE) == GLFW_PRESS)
        glfwSetWindowShouldClose(window, true);
}

// glfw: whenever the window size changed (by OS or user resize) this callback function executes
// ---------------------------------------------------------------------------------------------
void framebuffer_size_callback(GLFWwindow* window, int width, int height)
{
    // make sure the viewport matches the new window dimensions; note that width and 
    // height will be significantly larger than specified on retina displays.
    glViewport(0, 0, width, height);
}
