/*
 * JuiceFS, Copyright (C) 2020 Juicedata, Inc.
 *
 * This program is free software: you can use, redistribute, and/or modify
 * it under the terms of the GNU Affero General Public License, version 3
 * or later ("AGPL"), as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package meta

import (
	"syscall"
)

const (
	CHUNKBITS = 26 // 64M
	CHUNKSIZE = 1 << CHUNKBITS
	CHUNKMASK = CHUNKSIZE - 1

	CHUNK_DEL = 1000
)

const (
	TYPE_FILE      = 1
	TYPE_DIRECTORY = 2
	TYPE_SYMLINK   = 3
	TYPE_FIFO      = 4
	TYPE_BLOCKDEV  = 5
	TYPE_CHARDEV   = 6
	TYPE_SOCKET    = 7
)

const (
	SET_ATTR_MODE = 1 << iota
	SET_ATTR_UID
	SET_ATTR_GID
	SET_ATTR_SIZE
	SET_ATTR_ATIME
	SET_ATTR_MTIME
	SET_ATTR_CTIME
	SET_ATTR_ATIME_NOW
	SET_ATTR_MTIME_NOW

	// fallocate
	FALLOC_KEEP_SIZE      = 0x01
	FALLOC_PUNCH_HOLE     = 0x02
	FALLOC_NO_HIDE_STALE  = 0x04 // reserved
	FALLOC_COLLAPSE_RANGE = 0x08
	FALLOC_ZERO_RANGE     = 0x10
	FALLOC_INSERT_RANGE   = 0x20
)

type MsgCallback func(...interface{}) error

type Attr struct {
	Flags     uint8
	Typ       uint8
	Mode      uint16
	Uid       uint32
	Gid       uint32
	Atime     int64
	Mtime     int64
	Ctime     int64
	Atimensec uint32
	Mtimensec uint32
	Ctimensec uint32
	Nlink     uint32
	Length    uint64
	Rdev      uint32
	Full      bool
}

func typeToStatType(_type uint8) uint32 {
	switch _type & 0x7F {
	case TYPE_DIRECTORY:
		return syscall.S_IFDIR
	case TYPE_SYMLINK:
		return syscall.S_IFLNK
	case TYPE_FILE:
		return syscall.S_IFREG
	case TYPE_FIFO:
		return syscall.S_IFIFO
	case TYPE_SOCKET:
		return syscall.S_IFSOCK
	case TYPE_BLOCKDEV:
		return syscall.S_IFBLK
	case TYPE_CHARDEV:
		return syscall.S_IFCHR
	default:
		panic(_type)
	}
}

func (a Attr) SMode() uint32 {
	return typeToStatType(a.Typ) | uint32(a.Mode)
}

type Entry struct {
	Inode Ino
	Name  []byte
	Attr  *Attr
}

type Slice struct {
	Chunkid uint64
	Clen    uint32
	Off     uint32
	Len     uint32
}

const (
	// posix_locks.cmd:
	POSIX_LOCK_CMD_GET = 0
	POSIX_LOCK_CMD_SET = 1
	POSIX_LOCK_CMD_TRY = 2
	POSIX_LOCK_CMD_INT = 3

	// posix_locks.type:
	POSIX_LOCK_UNLCK   = 0
	POSIX_LOCK_RDLCK   = 1
	POSIX_LOCK_WRLCK   = 2
	POSIX_LOCK_INVALID = 3
)

type Meta interface {
	Init(format Format) error
	Load() (*Format, error)

	StatFS(ctx Context, totalspace, availspace, iused, iavail *uint64) syscall.Errno
	Access(ctx Context, inode Ino, modemask uint16) syscall.Errno
	Lookup(ctx Context, parent Ino, name string, inode *Ino, attr *Attr) syscall.Errno
	GetAttr(ctx Context, inode Ino, attr *Attr) syscall.Errno
	SetAttr(ctx Context, inode Ino, set uint16, sggidclearmode uint8, attr *Attr) syscall.Errno
	Truncate(ctx Context, inode Ino, flags uint8, attrlength uint64, attr *Attr) syscall.Errno
	Fallocate(ctx Context, inode Ino, mode uint8, off uint64, size uint64) syscall.Errno
	ReadLink(ctx Context, inode Ino, path *[]byte) syscall.Errno
	Symlink(ctx Context, parent Ino, name string, path string, inode *Ino, attr *Attr) syscall.Errno
	Mknod(ctx Context, parent Ino, name string, _type uint8, mode uint16, cumask uint16, rdev uint32, inode *Ino, attr *Attr) syscall.Errno
	Mkdir(ctx Context, parent Ino, name string, mode uint16, cumask uint16, copysgid uint8, inode *Ino, attr *Attr) syscall.Errno
	Unlink(ctx Context, parent Ino, name string) syscall.Errno
	Rmdir(ctx Context, parent Ino, name string) syscall.Errno
	Rename(ctx Context, parentSrc Ino, nameSrc string, parentDst Ino, nameDst string, inode *Ino, attr *Attr) syscall.Errno
	Link(ctx Context, inodeSrc, parent Ino, name string, attr *Attr) syscall.Errno
	Readdir(ctx Context, inode Ino, wantattr uint8, entries *[]*Entry) syscall.Errno
	Create(ctx Context, parent Ino, name string, mode uint16, cumask uint16, inode *Ino, attr *Attr) syscall.Errno
	Open(ctx Context, inode Ino, flags uint8, attr *Attr) syscall.Errno
	Close(ctx Context, inode Ino) syscall.Errno
	Read(inode Ino, indx uint32, chunks *[]Slice) syscall.Errno
	NewChunk(ctx Context, inode Ino, indx uint32, offset uint32, chunkid *uint64) syscall.Errno
	Write(ctx Context, inode Ino, indx uint32, off uint32, slice Slice) syscall.Errno

	GetXattr(ctx Context, inode Ino, name string, vbuff *[]byte) syscall.Errno
	ListXattr(ctx Context, inode Ino, dbuff *[]byte) syscall.Errno
	SetXattr(ctx Context, inode Ino, name string, value []byte) syscall.Errno
	RemoveXattr(ctx Context, inode Ino, name string) syscall.Errno
	Flock(ctx Context, inode Ino, owner uint64, ltype uint32, block bool) syscall.Errno
	Getlk(ctx Context, inode Ino, owner uint64, ltype *uint32, start, end *uint64, pid *uint32) syscall.Errno
	Setlk(ctx Context, inode Ino, owner uint64, block bool, ltype uint32, start, end uint64, pid uint32) syscall.Errno

	OnMsg(mtype uint32, cb MsgCallback)
}
