cmake_minimum_required(VERSION 3.30)
project(rm_lines_simple_renderer)

set(CMAKE_CXX_STANDARD 20)
include_directories(librm_lines/rm_lines/headers)

add_subdirectory(librm_lines)

add_executable(rm_lines_simple_renderer src/main.cpp)
target_link_libraries(rm_lines_simple_renderer rm_lines nlohmann_json::nlohmann_json)
target_link_options(rm_lines_simple_renderer PRIVATE
    -sMODULARIZE=1
    -sEXPORT_NAME=RmRenderer
    -sEXPORTED_RUNTIME_METHODS=['FS_writeFile','FS_readFile','callMain']
    -sINVOKE_RUN=0
    -sALLOW_MEMORY_GROWTH
)
