# Dockerfile for Go development of cybot
# Initially from https://github.com/manishrjain/godev

FROM ubuntu:trusty
MAINTAINER cyborgshadow <cyborgshadow@cyborgshadow.com>

# Install all of our needed things
RUN apt-get update && apt-get install -y --no-install-recommends \
bzr \
cmake \
curl \
g++ \
git \
make \
man-db \
mercurial \
ncurses-dev \
procps \
python-dev \
python-pip \
ssh \
sudo \
tmux \
unzip \
wget \
xz-utils \
&& rm -rf /var/lib/apt/lists/* \
&& pip install ipython \
&& git clone https://github.com/gmarik/Vundle.vim.git /root/.vim/bundle/Vundle.vim \
&& git clone https://github.com/Valloric/YouCompleteMe.git /root/.vim/bundle/YouCompleteMe \
&& git clone https://github.com/cyborgshadow/misc-scripts.git /root/tmux --depth 1 \
&& cd /root/tmux \
&& git checkout .tmux.conf \
&& git checkout git_reset \
&& mv git_reset /usr/local/bin/ \
&& mv .tmux.conf $HOME/ \
&& rm -rf /root/tmux \
&& cd /root/.vim/bundle/YouCompleteMe && git submodule update --init --recursive \
&& ./install.sh --clang-completer

# We build vim manually only because we need py2 and vim > 8.0 which isn't in any ubuntu repo I could find.
# I separated these from the installs above only because these are ONLY to build vim.
RUN apt-get update \
&& apt-get install -y --no-install-recommends libncurses5-dev \
libgnome2-dev \
libgnomeui-dev \
libgtk2.0-dev \
libatk1.0-dev \
libbonoboui2-dev \
libcairo2-dev \
libx11-dev \
libxpm-dev \
libxt-dev \
ruby-dev \
lua5.1 \
lua5.1-dev \
libperl-dev \
&& apt-get remove -y vim \
vim-runtime \
gvim \
vim-tiny \
vim-common \
vim-gui-common \
vim-nox \
&& git clone https://github.com/vim/vim.git /root/vim \
&& cd /root/vim \
&& ./configure --with-features=huge \
            --enable-multibyte \
            --enable-rubyinterp=yes \
            --enable-pythoninterp=yes \
            --with-python-config-dir=/usr/lib/python2.7/config \
            --enable-perlinterp=yes \
            --enable-luainterp=yes \
            --enable-gui=gtk2 --enable-cscope --prefix=/usr \
&& make VIMRUNTIMEDIR=/usr/share/vim/vim80 \
&& make install \
&& rm -rf /root/vim \
&& rm -rf /var/lib/apt/lists/*  

COPY bashrc /root/.bashrc

ENV GOVERSION 1.7

RUN curl -O https://storage.googleapis.com/golang/go$GOVERSION.linux-amd64.tar.gz && tar -C /usr/local -xzf go$GOVERSION.linux-amd64.tar.gz

# Setup our GOPATH (used for Go coding)
# Hardcode the homedir to /root
# Add GOPATH to $PATH (needed for some things in Go)
ENV GOPATH /go
ENV HOME /root
ENV PATH /go/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/cyborgshadow

RUN go version | grep $GOVERSION

# Install required Libraries
RUN go get -u github.com/BurntSushi/toml
RUN go get -u github.com/cyborgshadow/cybot
RUN go get -u github.com/inconshreveable/log15
RUN git config --global core.editor vim

COPY vimrc /root/.vimrc
RUN vim +PluginInstall +qall
RUN vim +GoInstallBinaries +qall

CMD ["tmux", "-u2"]

# To code in your local host, set your GOPATH ENV and run docker with docker run -itv $GOPATH/src:/go/src thisimage
# cd into the source directory and build code

