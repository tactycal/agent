### About

Package stubUtils provides wrapper functions for reading from file, writing
to file, executing a command and stub interface for unit testing.


A stubUtils package has predefined structs for mocking the wrapper functions
which has been mentioned above. All those sturcts implements ioStub interface
which will inform you through Tester interface if any of the passing
arguments to functions or their outputs will not match expected.


