file "/etc/profile.d/garden.sh" do
  mode 0755

  content <<-EOF
. /etc/profile.d/gvm.sh
gvm use go
EOF
end
