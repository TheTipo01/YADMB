package manager

import (
	"strings"
	"testing"
)

func TestIsValidUrl(t *testing.T) {
	if !IsValidURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ") {
		t.Error("String is supposed to be a valid URL.")
	}

	if IsValidURL("foof") {
		t.Error("String is not supposed to be a valid URL.")
	}
}

func TestFormatDuration(t *testing.T) {
	if d := FormatDuration(350); d != "05:50" {
		t.Errorf("Formatted duration is wrong. Excepted %s, got %s.", "05:50", d)
	}
}

func TestByteCountSI(t *testing.T) {
	if b := ByteCountSI(8192); b != "8.2 kB" {
		t.Errorf("Wrong size. Excepted %s, got %s.", "8.2 kB", b)
	}
}

func TestFormatLongMessage(t *testing.T) {
	const longString = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur rhoncus feugiat porta. Ut quis varius orci. Morbi facilisis aliquet nibh vitae molestie. Proin lacinia molestie mauris eget pharetra. Vestibulum non ex in nulla tincidunt eleifend. Vestibulum dapibus nisl a enim mattis commodo. Nulla nec turpis enim.Ut arcu odio, aliquet id ligula id, euismod vulputate magna. Vestibulum dignissim lectus sed tellus iaculis, eu vehicula velit venenatis. Suspendisse potenti. Sed sit amet dictum lacus, in tincidunt magna. Integer massa eros, iaculis sit amet quam sed, sodales malesuada ante. Aenean id mi nec est iaculis scelerisque eu at arcu. Sed eu lobortis augue, scelerisque eleifend quam. Sed fermentum nisl risus, id congue nulla pharetra in. Nullam sollicitudin nisi id nunc rutrum tincidunt. Aliquam in ante quis justo tincidunt porta. In dictum velit eget aliquet mattis. Duis tempus est diam, id molestie lectus ultrices at. Ut pulvinar malesuada ipsum quis tempus. Donec in mauris non dolor dictum tincidunt. Fusce feugiat mauris a ipsum aliquam, sit amet varius nunc ultricies. Aliquam erat volutpat.Phasellus malesuada rutrum cursus. Aenean sagittis tempus turpis, eu consectetur turpis congue nec. Duis hendrerit quam eget ipsum vehicula, in sagittis diam facilisis. Mauris ac tortor arcu. Cras commodo lorem non nunc iaculis, sit amet sollicitudin leo placerat. Morbi at elit eu nibh malesuada varius. Mauris sed pulvinar mi. Duis elementum pulvinar velit at pellentesque. Vestibulum quis tellus nisl. Phasellus non mauris quam. Etiam a facilisis risus, vel euismod tortor. Vestibulum rutrum ipsum quis hendrerit semper. Phasellus sed elit finibus velit egestas cursus. Curabitur vestibulum leo erat.In hendrerit ullamcorper odio sed gravida. Aliquam pharetra est non congue auctor. Integer convallis sagittis leo vel pulvinar. Aliquam ultricies id diam in tristique. Maecenas at massa ac risus placerat viverra. Vivamus tristique mattis lacus non ornare. Nullam sed lectus vel odio cursus placerat non sed risus. Morbi lectus nisl, aliquet non cursus quis, faucibus sed dui. Pellentesque non cursus arcu, et efficitur nibh. Vestibulum eu commodo felis.Pellentesque placerat leo in sem varius consectetur. Donec vel est augue. Quisque sagittis lorem ac elementum suscipit. Praesent nec ante mattis eros finibus facilisis sed eget ante. Donec elementum metus bibendum lacus sodales, semper efficitur orci pulvinar. Nulla accumsan ipsum at augue maximus pellentesque nec sed odio. Ut ullamcorper est neque, vel vulputate quam vulputate non. Donec sed nulla quis nunc blandit vulputate. Vivamus ultricies sem sit amet nunc molestie pellentesque. Nunc sed lobortis diam. Sed cursus nunc ac orci lobortis, et ultrices ex pretium. Vivamus vel diam erat. Etiam interdum ornare massa eu rutrum. Mauris ac malesuada ex. Sed sollicitudin orci nisl, vel porttitor tortor convallis vel.Integer finibus molestie dolor at varius. Vivamus sit amet nisi luctus, lobortis libero in, suscipit mauris. Fusce lobortis nisi ut felis rhoncus congue. Curabitur massa tortor, ultrices vel quam a, hendrerit facilisis nisl. Morbi vel augue odio. Nulla facilisi. Morbi at ex nisl. Maecenas urna nisi, volutpat vitae sagittis posuere, auctor non lacus. Cras interdum rutrum ullamcorper. Quisque non magna accumsan, molestie elit ac, volutpat erat. Morbi dapibus id nisl eu porttitor. Ut lacus ex, tincidunt non placerat eu, tempor non nunc. In at tortor dolor.In sagittis, ipsum vitae ultricies molestie, felis massa sollicitudin urna, et aliquam massa ipsum et elit. Donec ut mollis metus, sit amet volutpat neque. Nullam tempus, elit quis sollicitudin vehicula, ipsum sem commodo ante, eu fermentum leo arcu nec dolor. Pellentesque euismod neque nec faucibus ultrices. Phasellus faucibus, metus vitae bibendum sagittis, velit nisi suscipit elit, faucibus fermentum eros lectus a magna. In hac habitasse platea dictumst. Mauris ullamcorper justo eu iaculis sodales. Curabitur sapien risus, interdum id ligula non, rutrum ultricies tellus.Nulla eu nibh et nibh fringilla imperdiet. Nam convallis vestibulum libero, sed auctor erat dictum quis. Phasellus placerat turpis tristique bibendum vestibulum. Suspendisse et faucibus erat, vel luctus ex. In vitae efficitur lectus, sed scelerisque mi. Etiam mi ligula, elementum ac luctus ac, lacinia vel nulla. Aliquam non nisi sed ligula ornare accumsan. Aliquam sollicitudin rhoncus justo.Etiam in ultrices nunc, ac commodo urna. Nullam pulvinar tortor vitae leo interdum fringilla non vel nulla. Nunc hendrerit lectus velit, ut fringilla libero pulvinar at. Vestibulum convallis lacinia odio in egestas. Phasellus sit amet ipsum vel felis posuere euismod. Aliquam fringilla, est nec tincidunt tempus, arcu urna vulputate sapien, facilisis dictum lorem nunc sed nibh. Curabitur arcu libero, elementum nec erat id, sagittis malesuada lectus. Aliquam sed feugiat neque, et iaculis tellus. Nullam quis consequat tortor.Curabitur ut accumsan risus. Curabitur augue nunc, euismod quis ligula ac, rutrum bibendum odio. Curabitur pharetra, neque eu imperdiet vulputate, mi lectus sollicitudin leo, vel sollicitudin ante libero sit amet erat. Nullam tempor, ipsum eu tempor facilisis, elit enim aliquam sem, sed eleifend nibh dolor in sem. Cras dui tellus, euismod et dapibus at, auctor condimentum ex. Curabitur ut justo accumsan, maximus orci in, condimentum eros. Donec non quam eget arcu imperdiet mattis quis id ligula. Sed maximus at tortor pretium maximus. Aenean id sagittis elit. Proin justo ante, malesuada in elit id, faucibus consequat lacus. Nullam egestas ex sit amet arcu sollicitudin efficitur. Nulla facilisi. Nulla molestie suscipit purus aliquam pharetra. Fusce ac diam sed nulla ornare feugiat ut non mi. Aliquam urna mi, laoreet vitae malesuada a, venenatis in felis. "
	var generated string

	for _, s := range FormatLongMessage(strings.Split(longString, ". ")) {
		if len(s) > 2000 {
			t.Error("FormatLongMessage exceed the 2000 character mark.")
		}
		generated += s
	}

	generated = strings.ReplaceAll(strings.ReplaceAll(generated, "\n", ""), " ", "")
	if generated != strings.ReplaceAll(strings.ReplaceAll(longString, ". ", ""), " ", "") {
		t.Error("Generated string is different then formatted one!")
	}
}

func TestCleanURL(t *testing.T) {
	if link := CleanURL("https://youtu.be/dQw4w9WgXcQ?feature=shared"); link != "https://youtu.be/dQw4w9WgXcQ" {
		t.Error("CleanURL failed. Got", link)
	}
}
