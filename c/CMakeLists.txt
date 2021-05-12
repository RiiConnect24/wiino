cmake_minimum_required(VERSION 3.0)

project(WiiNo VERSION 1.0)

add_executable(wiino wiino.cpp)

install(FILES wiino DESTINATION bin)
