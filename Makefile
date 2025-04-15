version := $(shell grep 'const version = ' *.* | head -n 1 | awk -F'"' '{print $$2}')
arch := $(shell go env GOARCH)
debworkdir := /tmp/tgunpkg
$(info Version ${version}/${arch})
all:
	make -C plugin
clean:
	make -C plugin clean
	rm -rf *.deb ${debworkdir}
install:
	make -C plugin install
echo:
${debworkdir}/usr/local/bin/tgun:
	make -C plugin clean all install DESTDIR=${debworkdir}
deb: ${debworkdir}/usr/local/bin/tgun tgun_${version}_${arch}.deb
tgun_${version}_${arch}.deb: ${debworkdir}/usr/local/bin/tgun
	mkdir -p ${debworkdir}/DEBIAN
	cp tgun.control ${debworkdir}/DEBIAN/control
	sed -i "s/Version: X.Y.Z/Version: ${version}/g" ${debworkdir}/DEBIAN/control
	sed -i "s/Architecture: all/Architecture: ${arch}/g" ${debworkdir}/DEBIAN/control
	echo '#!/bin/bash' >> ${debworkdir}/DEBIAN/postinst
	echo 'ldconfig' >> ${debworkdir}/DEBIAN/postinst
	chmod +x ${debworkdir}/DEBIAN/postinst
	fakeroot dpkg-deb --root-owner-group --build ${debworkdir} $@
	rm -rf ${debworkdir}