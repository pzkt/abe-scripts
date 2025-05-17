include ./Makefile.common

.PHONY: all

CXXFLAGS := $(CXX11FLAGS)$(OS_CXXFLAGS) -pthread -Wall -g -O2 -DSSL_LIB_INIT -I${ZROOT}/deps/root/include -I${ZROOT}/root/include
LDFLAGS := -L${ZROOT}/deps/root/lib -L${ZROOT}/root/lib -L/usr/local/lib -Wl,-rpath=${ZROOT}/deps/root/lib
LIBS := -lcrypto -lrelic -lrelic_ec -lopenabe -lbenchmark -lbenchmark_main

all: waters11

waters11: waters11.o
	$(CXX) -o waters11 $(CXXFLAGS) $(LDFLAGS) waters11.cpp $(LIBS)

test:
	./waters11

clean:
	rm -rf *.o *.dSYM waters11
