# Latest debian, with curl, should contain all necessary dependencies
# https://github.com/devcontainers/images/tree/main/src/base-debian
# FROM mcr.microsoft.com/devcontainers/base:debian
# https://github.com/docker-library/buildpack-deps/blob/98a5ab81d47a106c458cdf90733df0ee8beea06c/debian/buster/curl/Dockerfile
FROM buildpack-deps:curl 

# Update and install necessary system packages
RUN apt-get update \
    && apt-get install -y \
        sudo \
        xz-utils \
        tar \
        curl \
        procps \
        zsh

# Create a non-root user
# from https://code.visualstudio.com/remote/advancedcontainers/add-nonroot-user#_creating-a-nonroot-user
# alternatively, could use this feature to do it automagically at boot time
#  https://github.com/devcontainers/features/blob/main/src/common-utils/main.sh#L371
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
# Create the user with sudo privileges
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME --shell /bin/zsh \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Install nix with flakes and the unified-command enabled
RUN curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix > /nix-installer && chmod +x /nix-installer
RUN /nix-installer install linux --init none --no-confirm

# Add user to trusted users
RUN echo "extra-trusted-users = $USERNAME" | sudo tee -a /etc/nix/nix.conf >/dev/null
# Allow the user to use the nix daemon without sudo
RUN usermod -a -G nixbld $USERNAME

# Initialize nix
COPY nixinstall.sh /tmp/nixinstall.sh
RUN chmod +x /tmp/nixinstall.sh && /tmp/nixinstall.sh && rm /tmp/nixinstall.sh

# Add the entrypoint
COPY entrypoint.sh /tmp/entrypoint.sh
RUN install --mode 755 /tmp/entrypoint.sh /usr/local/bin/entrypoint.sh && rm /tmp/entrypoint.sh

COPY update-content.sh /tmp/update-content.sh
RUN install --mode 755 /tmp/update-content.sh /usr/local/bin/update-content.sh && rm /tmp/update-content.sh

# Make this bash wrapper the default source for all scripts
COPY sh.sh /tmp/sh.sh
RUN rm /bin/sh && install --mode 755 /tmp/sh.sh /bin/sh && rm /tmp/sh.sh

# Enable direnv on all directories by default
USER $USERNAME
RUN mkdir -p /home/$USERNAME/.config/direnv/
COPY direnv.toml /home/$USERNAME/.config/direnv/direnv.toml
# setup direnv and starship bash and zsh
RUN touch ~/.zshrc ~/.bashrc
RUN echo 'eval "$(direnv hook zsh)"' >> /home/$USERNAME/.zshenv
RUN echo 'eval "$(starship init zsh)"' >> /home/$USERNAME/.zshenv
RUN echo 'eval "$(direnv hook bash)"' >> /home/$USERNAME/.bash_aliases
RUN echo 'eval "$(starship init bash)"' >> /home/$USERNAME/.bash_aliases

# By default, starting the container will run forever. This is necessary for
# some reason that I don't understand.
ENTRYPOINT [ "entrypoint.sh" ]
CMD ["sleep", "infinity"]
