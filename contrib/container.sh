BASE=honeytrap

echo nameserver 8.8.8.8 > /var/lib/lxc/$BASE/rootfs/etc/resolv.conf
echo nameserver 8.8.4.4 >> /var/lib/lxc/$BASE/rootfs/etc/resolv.conf

echo "lxc.console = none" >> /var/lib/lxc/$BASE/config
echo "lxc.tty = 0" >> /var/lib/lxc/$BASE/config
echo "lxc.cgroup.devices.deny = c 5:1 rwm" >> /var/lib/lxc/$BASE/config
echo "lxc.cgroup.memory.limit_in_bytes = 512m" >> /var/lib/lxc/$BASE/config

HOSTNAME=$(curl https://gist.githubusercontent.com/nl5887/41ec1a4aa38bd6715f69/raw/servernames| shuf -n 0)
echo $HOSTNAME > /var/lib/lxc/$BASE/rootfs/etc/hostname

chroot /var/lib/lxc/$BASE/rootfs apt-get update -y
chroot /var/lib/lxc/$BASE/rootfs apt-get upgrade -y
chroot /var/lib/lxc/$BASE/rootfs apt-get install -y openssh-server curl wget apache2
chroot /var/lib/lxc/$BASE/rootfs sed -i.bak 's/^PermitRootLogin without-password/PermitRootLogin yes/g' /etc/ssh/sshd_config

echo 'root:root' | chroot /var/lib/lxc/$BASE/rootfs chpasswd
echo 'ubuntu:ubuntu' | chroot /var/lib/lxc/$BASE/rootfs chpasswd
