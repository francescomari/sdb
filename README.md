# Segment Debugger

The Segment Debugger is a collection of utilities to inspect a repository created by the [Segment Store](http://jackrabbit.apache.org/oak/docs/nodestore/segment/overview.html), a component in [Apache Jackrabbit Oak](http://jackrabbit.apache.org/oak/).

## Installing

To start using the Segment Debugger, install Go and run 'go get' as shown below.

```
$ go get -u github.com/francescomari/sdb
```

This will retrieve the library and install the `sdb` command line utility into your `$GOBIN` path.

## List TAR files

The `tars` command can be used to list TAR files in a specific folder.

```
$ sdb tars store
data00000b.tar
data00001a.tar
```

TAR files can exits in multiple generations.
By default, the `tars` command shows only the most recent generation of the TAR files.
You can use the `tars` command to list every TAR file in a folder by specifying the `-all` flag.

```
$ sdb tars -all store
data00000a.tar
data00000b.tar
data00001a.tar
```

You don't have to specify the folder on the command line.
If you don't specify the folder, the current working directory is assumed.

```
$ cd store && sdb tars
data00000b.tar
data00001a.tar
```

## List entries in a TAR file

The `entries` command lists the name of the entries in a TAR file, in the same order as they appear in the file.

```
$ sdb entries data00000a.tar | tail
852365fa-9a90-442d-b3cc-382d0615beff.77784815
dd789211-2515-4d34-bba4-5b179b7b794d.90ff266d
613ba5f3-01b3-4571-ba90-64a182d7e7c7.269e8f6b
b92806c8-e895-4a33-bcec-3d12a7262963.6e679142
647cbafa-ac92-4078-b1fd-4a3891216593.e3d9c6ab
5438284c-2280-4dea-b462-cac52f18039d.00780c6d
541f0c4c-b04c-4da8-b5d8-aa29a74c95fe.477d84f5
data00000a.tar.brf
data00000a.tar.gph
data00000a.tar.idx
```

The output below shows that a TAR files produced by the Segment Store is a collection of segment entries and is always terminated by some entries containing metadata about the segments.

## List segment IDs in a TAR file

The `segments` command lists the segment ID associated to every segment entry in a TAR file.
The segment IDs are presented in the same order as their corresponding entries in the TAR file.

```
$ sdb segments data00000a.tar | head
data 0ce1d7f06f464753a42c2374852990c8
data cc505b3d7568419faae2f8b4311911a9
data 2a2db8b66348406aadf309bc1a471b0a
data b0540a20c4e14f41a62d1eb2ae065b86
data 81386db441b441dea59de28d787088ff
data 082729868cd84c74a03b1014d57f242c
bulk 0b400d31c52d43a4b10cb925fb5f9d14
bulk 54527edc88204188b4000805e8c74e66
bulk ae6ce7066599406fb5c0332581f07989
bulk 4ab153cf35c84901bd5c1b74acc2bd0b
```

Segments can either be data or bulk segments.
The first column of the output allows you to easily distinguish between the two types.
Moreover, having the segment type spelled out comes in handy when searching for only a specific kind of segment.

```
$ sdb segments data00000a.tar | grep data
data 0ce1d7f06f464753a42c2374852990c8
data cc505b3d7568419faae2f8b4311911a9
data 2a2db8b66348406aadf309bc1a471b0a
data b0540a20c4e14f41a62d1eb2ae065b86
data 81386db441b441dea59de28d787088ff
data 082729868cd84c74a03b1014d57f242c
data 8b1123b2ba1442acad869ba1bafdbd5c
data ad2ec6b0eaac439cad5eb9277082bcb3
data c73aaf84b38c4498afe57cd86d4d8e62
data 4e0627f5a3f54929a531afe5be4b4a36
data e53c50281d114feba88c124007fbb5e9
data 4e815f3e9b23429aa0ee4b967c7566c1
```

## Show the content of a segment

The `segment` command shows you the hexdump of a segment.
You need to specify the TAR file the segment belongs to and its ID.
It is possible to access a hexdump of the segment by using the `-format` flag.

```
$ sdb segment data00000a.tar 0ce1d7f06f464753a42c2374852990c8
version 12
generation 0
reference 1 9bfa18e9bbd04ae2ab00451f185b17fe
reference 2 195aa442cfbc4fbea1157288e94763ad
...
reference 27 c3857c61941044aba69df3e35c519728
record 0 value 3ffd0
record 1 bucket 3ffa0
...
record 16 node 3fcd8
```

The output shows a readable representation of the fields of the segment.
The name of the field is the first item printed on every line.
The following fields are supported:
* `version`
The version of the segment.
* `generation`
The generation this segment belongs to.
* `reference`
A multi-value field containing references to other segments.
For every reference, its number and the corresponding segment is shown.
* `record`
A multi-value field containing pointers to the records in this segment.
For every record, its logical number, type and offset is shown.
The offset of the record is unnormalized and relative from the end of the segment.
The type of the record is a string that can assume the values `block`, `list`, `bucket`, `branch`, `leaf`, `node`, `template`, `value`, `binary` and `unknown`.

## Show the content of the index

The `index` command prints the content of the TAR index.
It is possible to access a hexdump of the index by using the `-format` flag.

```
$ sdb index data00000a.tar | head
data 80120c8f7d984b81a9cda2b4aeaf5686 5de8a00 262112 0
data 8017fadcca4e4398a92efb64f0c41b19 241800 262096 0
data 828439805d274f4ea982fd8e93bcc1d1 4fa8e00 250192 0
data 83019b303a8a4c1aa20cbad77aab559a 54ffc00 262112 0
data 83c7fcb87cdf428faa2838b8d81dc2a3 51a5c00 262112 0
data 83e08e3c8eba497bacdb9c6fb10c8785 59aa800 261072 0
data 84703034cf834936ad206d11b9773a3d 2ddd800 257648 0
data 85837008c5624164a6268d76b00f6bab 1c9a800 258160 0
data 8633a21403c44cd0a026ec233cc683be 438da00 262096 0
data 86a132a870cc4a83addf449c4ec404b8 4904800 262096 0
```

The output shows the following columns: the type of the segment, the segment ID, the hexadecimal offset of the segment in the TAR file, the size of the segment, and its generation.

## Show the content of the graph

The `graph` command prints the content of the TAR graph.
It is possible to access a hexdump of the graph by using the `-format` flag.

```
$ sdb graph data00000a.tar | head -n 5
4535f3ee3bb543f5a682f9b64e5d8bf2 6c98954462fa4bd7ab50a15f064f864d
4535f3ee3bb543f5a682f9b64e5d8bf2 d012d6f392814ba5ad6bc0c3013ce12e
16ae8fb02f0a4e0faa49a281e98d8d5e 4535f3ee3bb543f5a682f9b64e5d8bf2
16ae8fb02f0a4e0faa49a281e98d8d5e 94bdb06b064c4f5ba13566a257ba99ef
16ae8fb02f0a4e0faa49a281e98d8d5e ca615810fd7b4a7ca33c9b1a76b8fbeb
```

The output shows a graph represented as an adjacency list.
In the output above, the first two lines show that segment `4535f3ee...` has two edges directed to the segments `6c989544...`  and `d012d6f3...`.
The following lines show three edges directed from segment `16ae8fb0..` towards segments `4535f3ee...`, `94bdb06b...` and `ca615810`.

## Show the content of the binary references index

The `binaries` command prints the content of the binary references index of a TAR file.
It is possible to access a hexdump of the binary references index by using the `-format` flag.

```
$ sdb binaries -format text data00000a.tar | head -n 5
0 12c552d1d67f4b4fa22a61c5818286a2 f20cc9f7902d6facdd7a9e260dc686d144de5ca3#108232
0 12c552d1d67f4b4fa22a61c5818286a2 4ab8c9485e1c13410eb684863f333414e0e2973d#37470
0 45ef53df2d094fcea8336b1fdc7c3e49 360636479c1c3b1b47d2174d4432224fc6193ed4#22090
0 45ef53df2d094fcea8336b1fdc7c3e49 9e56c46ff9b64986f491c53e0f625fa1fd26daff#84533
0 5666923c93a54d21a9a36cb3372a890b dab128f67d237c980e5770ee045c38e36afa1327#19764
```

The output shows a list of external binary references.
Every line contains the generation of the segment, the segment ID and the identifier of the binary reference.
The example above shows only segments of generation `0`.
One of those segments is `12c552d1...`.
This segment has two references to the binaries identified by `f20cc9f7...` and `4ab8c948...`.
