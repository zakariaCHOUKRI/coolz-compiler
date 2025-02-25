target triple = "x86_64-pc-windows-msvc19.43.34808"

@str.0 = global [55 x i8] c"Error: the program was aborted by an abort() function\0A\00"
@str.1 = global [7 x i8] c"Object\00"
@str.2 = global [3 x i8] c"%s\00"
@str.3 = global [5 x i8] c"%lld\00"
@str.4 = global [9 x i8] c"%255[^\0A]\00"
@str.5 = global [4 x i8] c"%*c\00"
@str.6 = global [28 x i8] c"Error: substr out of range\0A\00"
@str.7 = global [5 x i8] c"Main\00"
@str.8 = global [25 x i8] c"Please enter your name: \00"
@str.9 = global [15 x i8] c"Your name is: \00"
@str.10 = global [2 x i8] c"\0A\00"
@str.11 = global [24 x i8] c"Please enter a number: \00"
@str.12 = global [17 x i8] c"Your number is: \00"

declare i32 @printf(i8* %format, ...)

declare i32 @scanf(i8* %format, ...)

declare i8* @memset(i8* %str, i32 %c, i64 %n)

declare i8* @malloc(i64 %size)

declare i8* @memcpy(i8* %dest, i8* %src, i64 %size)

declare i64 @strlen(i8* %str)

define i8* @Object_abort(i8* %self) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([55 x i8], [55 x i8]* @str.0, i32 0, i32 0))
	call void @exit(i32 1)
	unreachable
}

declare void @exit(i32 %status)

define i8* @Object_type_name(i8* %self) {
0:
	ret i8* getelementptr ([7 x i8], [7 x i8]* @str.1, i32 0, i32 0)
}

define i8* @Object_copy(i8* %self) {
0:
	%1 = call i8* @malloc(i64 8)
	%2 = call i8* @memcpy(i8* %1, i8* %self, i64 8)
	ret i8* %1
}

define i8* @IO_out_string(i8* %self, i8* %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([3 x i8], [3 x i8]* @str.2, i32 0, i32 0), i8* %x)
	ret i8* %self
}

define i8* @IO_out_int(i8* %self, i64 %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([5 x i8], [5 x i8]* @str.3, i32 0, i32 0), i64 %x)
	ret i8* %self
}

define i8* @IO_in_string(i8* %self) {
0:
	%1 = alloca [256 x i8]
	%2 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%3 = call i8* @memset(i8* %2, i32 0, i64 256)
	%4 = call i32 (i8*, ...) @scanf(i8* getelementptr ([9 x i8], [9 x i8]* @str.4, i32 0, i32 0), [256 x i8]* %1)
	%5 = call i32 (i8*, ...) @scanf(i8* getelementptr ([4 x i8], [4 x i8]* @str.5, i32 0, i32 0))
	%6 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%7 = call i64 @strlen(i8* %6)
	%8 = add i64 %7, 1
	%9 = call i8* @malloc(i64 %8)
	%10 = call i8* @memcpy(i8* %9, i8* %6, i64 %8)
	ret i8* %9
}

define i64 @IO_in_int(i8* %self) {
0:
	%1 = alloca i64
	%2 = call i32 (i8*, ...) @scanf([5 x i8]* @str.3, i64* %1)
	%3 = call i32 (i8*, ...) @scanf([4 x i8]* @str.5)
	%4 = load i64, i64* %1
	ret i64 %4
}

define i64 @String_length(i8* %self) {
0:
	%1 = call i64 @strlen(i8* %self)
	ret i64 %1
}

define i8* @String_substr(i8* %self, i64 %i, i64 %l) {
0:
	%1 = call i64 @strlen(i8* %self)
	%2 = icmp slt i64 %i, 0
	%3 = add i64 %i, %l
	%4 = icmp sgt i64 %3, %1
	%5 = icmp slt i64 %l, 0
	%6 = or i1 %2, %4
	%7 = or i1 %6, %5
	br i1 %7, label %error, label %success

error:
	%8 = call i32 (i8*, ...) @printf(i8* getelementptr ([28 x i8], [28 x i8]* @str.6, i32 0, i32 0))
	%9 = call i8* @Object_abort(i8* %self)
	unreachable

success:
	%10 = getelementptr i8, i8* %self, i64 %i
	%11 = add i64 %l, 1
	%12 = call i8* @malloc(i64 %11)
	%13 = call i8* @memcpy(i8* %12, i8* %10, i64 %l)
	%14 = getelementptr i8, i8* %12, i64 %l
	store i8 0, i8* %14
	ret i8* %12
}

define i8* @String_concat(i8* %self, i8* %s) {
0:
	%1 = call i64 @strlen(i8* %self)
	%2 = call i64 @strlen(i8* %s)
	%3 = add i64 %1, %2
	%4 = add i64 %3, 1
	%5 = call i8* @malloc(i64 %4)
	%6 = call i8* @memcpy(i8* %5, i8* %self, i64 %1)
	%7 = getelementptr i8, i8* %5, i64 %1
	%8 = add i64 %2, 1
	%9 = call i8* @memcpy(i8* %7, i8* %s, i64 %8)
	ret i8* %5
}

define i8* @Main_main(i8* %self) {
0:
	%1 = call i8* @IO_out_string(i8* %self, i8* getelementptr ([25 x i8], [25 x i8]* @str.8, i32 0, i32 0))
	%2 = alloca i8*
	%3 = call i8* @IO_in_string(i8* %self)
	store i8* %3, i8** %2
	%4 = call i8* @IO_out_string(i8* %self, i8* getelementptr ([15 x i8], [15 x i8]* @str.9, i32 0, i32 0))
	%5 = load i8*, i8** %2
	%6 = call i8* @IO_out_string(i8* %self, i8* %5)
	%7 = call i8* @IO_out_string(i8* %self, i8* getelementptr ([2 x i8], [2 x i8]* @str.10, i32 0, i32 0))
	%8 = call i8* @IO_out_string(i8* %self, i8* getelementptr ([24 x i8], [24 x i8]* @str.11, i32 0, i32 0))
	%9 = alloca i64
	%10 = call i64 @IO_in_int(i8* %self)
	store i64 %10, i64* %9
	%11 = call i8* @IO_out_string(i8* %self, i8* getelementptr ([17 x i8], [17 x i8]* @str.12, i32 0, i32 0))
	%12 = load i64, i64* %9
	%13 = call i8* @IO_out_int(i8* %self, i64 %12)
	%14 = call i8* @IO_out_string(i8* %self, [2 x i8]* @str.10)
	%15 = call i8* @IO_out_string(i8* %self, [24 x i8]* @str.11)
	%16 = alloca i64
	%17 = call i64 @IO_in_int(i8* %self)
	store i64 %17, i64* %16
	%18 = call i8* @IO_out_string(i8* %self, [17 x i8]* @str.12)
	%19 = load i64, i64* %16
	%20 = call i8* @IO_out_int(i8* %self, i64 %19)
	%21 = call i8* @IO_out_string(i8* %self, [2 x i8]* @str.10)
	%22 = call i8* @IO_out_string(i8* %self, [25 x i8]* @str.8)
	%23 = alloca i8*
	%24 = call i8* @IO_in_string(i8* %self)
	store i8* %24, i8** %23
	%25 = call i8* @IO_out_string(i8* %self, [15 x i8]* @str.9)
	%26 = load i8*, i8** %23
	%27 = call i8* @IO_out_string(i8* %self, i8* %26)
	%28 = call i8* @IO_out_string(i8* %self, [2 x i8]* @str.10)
	ret i8* %28
}

define i8* @Main_type_name(i8* %self) {
0:
	ret i8* getelementptr ([5 x i8], [5 x i8]* @str.7, i32 0, i32 0)
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
