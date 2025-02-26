target triple = "x86_64-pc-windows-msvc19.43.34808"

@str.0 = global [55 x i8] c"Error: the program was aborted by an abort() function\0A\00"
@str.1 = global [7 x i8] c"Object\00"
@str.2 = global [3 x i8] c"%s\00"
@str.3 = global [5 x i8] c"%lld\00"
@str.4 = global [9 x i8] c"%255[^\0A]\00"
@str.5 = global [4 x i8] c"%*c\00"
@str.6 = global [28 x i8] c"Error: substr out of range\0A\00"
@str.7 = global [8 x i8] c"Printer\00"
@str.8 = global [2 x i8] c"\0A\00"
@str.9 = global [5 x i8] c"Main\00"
@str.10 = global [15 x i8] c"Hello everyone\00"
@str.11 = global [21 x i8] c"testing length() ...\00"
@str.12 = global [15 x i8] c"the length of:\00"
@str.13 = global [2 x i8] c"\22\00"
@str.14 = global [3 x i8] c"\22 \00"
@str.15 = global [5 x i8] c"is: \00"
@str.16 = global [1 x i8] c"\00"
@str.17 = global [11 x i8] c"0123456789\00"
@str.18 = global [21 x i8] c"testing substr() ...\00"
@str.19 = global [59 x i8] c"substring index from 3 to 7 of the string \220123456789\22 is:\00"
@str.20 = global [26 x i8] c"Please input a string a: \00"
@str.21 = global [26 x i8] c"Please input a string b: \00"
@str.22 = global [34 x i8] c"The concatenation of a and b is: \00"

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

define i8* @Printer_print(i8* %self, i8* %x) {
0:
	%1 = alloca i8*
	store i8* %x, i8** %1
	%2 = load i8*, i8** %1
	%3 = call i8* @IO_out_string(i8* %self, i8* %2)
	ret i8* %3
}

define i8* @Printer_println(i8* %self, i8* %x) {
0:
	%1 = alloca i8*
	store i8* %x, i8** %1
	%2 = load i8*, i8** %1
	%3 = call i8* @Printer_print(i8* %self, i8* %2)
	%4 = call i8* @Printer_print(i8* %self, i8* getelementptr ([2 x i8], [2 x i8]* @str.8, i32 0, i32 0))
	ret i8* %4
}

define i8* @Printer_type_name(i8* %self) {
0:
	ret i8* getelementptr ([8 x i8], [8 x i8]* @str.7, i32 0, i32 0)
}

define i8* @Main_main(i8* %self) {
0:
	%1 = alloca i8*
	store i8* getelementptr ([15 x i8], [15 x i8]* @str.10, i32 0, i32 0), i8** %1
	%2 = call i8* @Printer_println(i8* %self, i8* getelementptr ([21 x i8], [21 x i8]* @str.11, i32 0, i32 0))
	%3 = call i8* @Printer_println(i8* %self, i8* getelementptr ([15 x i8], [15 x i8]* @str.12, i32 0, i32 0))
	%4 = call i8* @Printer_print(i8* %self, i8* getelementptr ([2 x i8], [2 x i8]* @str.13, i32 0, i32 0))
	%5 = load i8*, i8** %1
	%6 = call i8* @Printer_print(i8* %self, i8* %5)
	%7 = call i8* @Printer_print(i8* %self, i8* getelementptr ([3 x i8], [3 x i8]* @str.14, i32 0, i32 0))
	%8 = call i8* @Printer_print(i8* %self, i8* getelementptr ([5 x i8], [5 x i8]* @str.15, i32 0, i32 0))
	%9 = load i8*, i8** %1
	%10 = call i64 @String_length(i8* %9)
	%11 = call i8* @IO_out_int(i8* %self, i64 %10)
	%12 = call i8* @Printer_println(i8* %self, i8* getelementptr ([1 x i8], [1 x i8]* @str.16, i32 0, i32 0))
	%13 = call i8* @Printer_println(i8* %self, [1 x i8]* @str.16)
	%14 = alloca i8*
	store i8* getelementptr ([11 x i8], [11 x i8]* @str.17, i32 0, i32 0), i8** %14
	%15 = call i8* @Printer_println(i8* %self, i8* getelementptr ([21 x i8], [21 x i8]* @str.18, i32 0, i32 0))
	%16 = call i8* @Printer_println(i8* %self, i8* getelementptr ([59 x i8], [59 x i8]* @str.19, i32 0, i32 0))
	%17 = load i8*, i8** %14
	%18 = sub i64 7, 3
	%19 = add i64 %18, 1
	%20 = call i8* @String_substr(i8* %17, i64 3, i64 %19)
	%21 = call i8* @Printer_println(i8* %self, i8* %20)
	%22 = call i8* @Printer_println(i8* %self, [1 x i8]* @str.16)
	%23 = call i8* @Printer_print(i8* %self, i8* getelementptr ([26 x i8], [26 x i8]* @str.20, i32 0, i32 0))
	%24 = alloca i8*
	%25 = call i8* @IO_in_string(i8* %self)
	store i8* %25, i8** %24
	%26 = call i8* @Printer_print(i8* %self, i8* getelementptr ([26 x i8], [26 x i8]* @str.21, i32 0, i32 0))
	%27 = alloca i8*
	%28 = call i8* @IO_in_string(i8* %self)
	store i8* %28, i8** %27
	%29 = alloca i8*
	%30 = load i8*, i8** %24
	%31 = load i8*, i8** %27
	%32 = call i8* @String_concat(i8* %30, i8* %31)
	store i8* %32, i8** %29
	%33 = call i8* @Printer_println(i8* %self, i8* getelementptr ([34 x i8], [34 x i8]* @str.22, i32 0, i32 0))
	%34 = load i8*, i8** %29
	%35 = call i8* @Printer_println(i8* %self, i8* %34)
	ret i8* %35
}

define i8* @Main_type_name(i8* %self) {
0:
	ret i8* getelementptr ([5 x i8], [5 x i8]* @str.9, i32 0, i32 0)
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
