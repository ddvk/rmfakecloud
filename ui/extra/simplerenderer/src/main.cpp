#include "library.h"
#include "renderer/renderer_export.h"
#include "scene_tree/scene_tree_export.h"
#include <fstream>
#include <iostream>
#include <string>
#include <cstdint>

static inline void write32BitInt(std::ostream &ostr, uint32_t i) {
    uint8_t data[4] = {
        (uint8_t) ((i >> 24) & 0xFF),
        (uint8_t) ((i >> 16) & 0xFF),
        (uint8_t) ((i >> 8) & 0xFF),
        (uint8_t) ((i >> 0) & 0xFF),
    };
    ostr.write((char*) data, sizeof(data));
}

int main(int argc, char **argv){
    if(argc != 3) {
        std::cerr << "Usage: " << *argv << " <input .rm file> <output raw bitmap buffer>" << std::endl;
        return -2; 
    }
    std::string in(argv[1]);
    std::string out(argv[2]);

    // Render the file:
    const char *treeId = buildTree(in.c_str());
    if(!treeId) {
        std::cerr << "Failed to build the tree!" << std::endl;
        return -1;
    }

    auto tree = getSceneTree(treeId);
    std::optional<Renderer> renderer;
    try {
        renderer = Renderer(tree.get(), NOTEBOOK, false);
    } catch (const std::exception &e) {
        std::cerr << "Failed to create renderer!" << std::endl;
        destroyTree(treeId);
        return -1;
    }
    std::ofstream outputFile(out.c_str());

    Rect track(0, 0, 0, 0);
    for(auto &ref : renderer->layers) {
        auto layer = renderer->getSizeTracker(ref.groupId);
        if (layer->getBottom() > track.getBottom()) {
            track.setBottom(layer->getBottom());
        }
        if (layer->getTop() < track.getTop()) {
            track.setTop(layer->getTop());
        }
        if (layer->getLeft() < track.getLeft()) {
            track.setLeft(layer->getLeft());
        }
        if (layer->getRight() > track.getRight()) {
            track.setRight(layer->getRight());
        }
    }
    uint32_t width = std::max((uint32_t) (track.getRight() - track.getLeft()), renderer->paperSize.first);
    uint32_t height = std::max((uint32_t) (track.getBottom() - track.getTop()), renderer->paperSize.second);
    uint32_t size = width * height * 4;
    uint32_t *rawFrame = (uint32_t*) malloc(size);
    renderer->getFrame(rawFrame, size, Vector(track.getLeft(), track.getTop()), Vector(width, height), Vector(width, height), true);
    write32BitInt(outputFile, width);
    write32BitInt(outputFile, height);
    outputFile.write((char *) rawFrame, size);
    destroyTree(treeId);
}
