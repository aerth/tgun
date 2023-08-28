TEMPLATE = app
TARGET = plugin
INCLUDEPATH += ..
LIBS += -L../ $$PWD/../tgun.so

#DEFINES += QT_DISABLE_DEPRECATED_BEFORE=0x060000    # disables all the APIs deprecated before Qt 6.0.0

# Input
SOURCES += example_tcp.cpp
